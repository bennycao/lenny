package main

import (
	"bytes"
	"fmt"

	"github.com/jasonlvhit/gocron"
	"github.com/nlopes/slack"
	"github.com/zmb3/spotify"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

type BotCentral struct {
	Channel *slack.Channel
	Event   *slack.MessageEvent
	UserId  string
}

type ReplyChannel struct {
	Channel      *slack.Channel
	Attachments  []slack.Attachment
	DisplayTitle string
}

type StandupChannel struct {
	Channel     *slack.Channel
	StandupTime time.Time
}

var (
	commandChannel chan *BotCentral
	replyChannel   chan ReplyChannel
	standupChannel chan StandupChannel
	api            *slack.Client
	standupTime    time.Time
	genresSeeds    []string
	setTimes       []string
)

func startBot() {
	token := os.Getenv("SLACK_TOKEN")
	api = slack.New(token)
	rtm := api.NewRTM()

	commandChannel = make(chan *BotCentral)
	replyChannel = make(chan ReplyChannel)
	standupChannel = make(chan StandupChannel)

	go rtm.ManageConnection()
	go handleCommands(replyChannel, standupChannel)
	go handleReply()
	go handleStandupTimer(replyChannel)

Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				fmt.Println("Connection counter:", ev.ConnectionCount)

			case *slack.MessageEvent:
				channelInfo, err := api.GetChannelInfo(ev.Channel)
				if err != nil {
					log.Println(err)
				}

				botCentral := &BotCentral{
					Channel: channelInfo,
					Event:   ev,
					UserId:  ev.User,
				}

				if ev.Type == "message" && strings.HasPrefix(ev.Text, "<@"+rtm.GetInfo().User.ID+">") {
					commandChannel <- botCentral
				}

			case *slack.RTMError:
				fmt.Printf("Error: %s\n", ev.Error())

			case *slack.InvalidAuthEvent:
				fmt.Printf("Invalid credentials")
				break Loop

			default:
				//Take no action
			}
		}
	}
}

func handleCommands(c chan ReplyChannel, sc chan StandupChannel) {
	commands := map[string]string{
		"help":      "see the available commands",
		"play":      "play some tunes LENNY !",
		"stop":      "pause|stop da music",
		"next":      "next tune LENNY, this ain't any good",
		"previous":  "dat tune was ace, play it again LENNY",
		"add":       "queue dis beats",
		"list":      "list me current tunes",
		"set-time":  "When's my set time ?",
		"set-times": "Give me the set times",
		"genres":    "List me some flavours",
		"current":   "What's currently on da radio?",
		"search":    "Where are all the tunes ?",
	}

	var replyChannel ReplyChannel
	var standupChannel StandupChannel

	for {
		botChannel := <-commandChannel
		replyChannel.Channel = botChannel.Channel
		standupChannel.Channel = botChannel.Channel
		commandArray := strings.Fields(botChannel.Event.Text)
		log.Printf("%+v\n", commandArray)
		switch commandArray[1] {
		case "help":
			fields := make([]slack.AttachmentField, 0)
			for k, v := range commands {
				fields = append(fields, slack.AttachmentField{
					Title: "@lenny " + k,
					Value: v,
				})
			}
			attachment := slack.Attachment{
				Pretext: "Rockn' tune commands",
				Color:   "#85929E",
				Fields:  fields,
			}
			replyChannel.Attachments = []slack.Attachment{attachment}
			c <- replyChannel
		case "play":

			client.Play()
			time.Sleep(2)
			cp, _ := client.PlayerCurrentlyPlaying()

			attachment := slack.Attachment{
				Pretext: cp.Item.Name,
				Color:   "#85929E",
			}
			replyChannel.Attachments = []slack.Attachment{attachment}
			c <- replyChannel
		case "stop":
			client.Pause()
			attachment := slack.Attachment{
				Pretext: "Pausing da tunes",
				Color:   "#85929E",
			}
			replyChannel.Attachments = []slack.Attachment{attachment}
			c <- replyChannel
		case "next":
			client.Next()
			time.Sleep(2 * time.Second)
			cp, _ := client.PlayerCurrentlyPlaying()

			attachment := slack.Attachment{
				Pretext: cp.Item.Name,
				Color:   "#85929E",
			}
			replyChannel.Attachments = []slack.Attachment{attachment}
			c <- replyChannel
		case "previous":
			client.Previous()
			time.Sleep(2 * time.Second)
			cp, _ := client.PlayerCurrentlyPlaying()

			attachment := slack.Attachment{
				Pretext: cp.Item.Name,
				Color:   "#85929E",
			}
			replyChannel.Attachments = []slack.Attachment{attachment}
			c <- replyChannel

		case "search":
			//fields := make([]slack.AttachmentField, 0)
			attachments := make([]slack.Attachment, 0)
			var buffer bytes.Buffer

			for i := 2; i < len(commandArray); i++ {
				buffer.WriteString(commandArray[i])
				buffer.WriteString(" ")
			}

			r, _ := client.Search(buffer.String(), spotify.SearchTypeTrack)

			if r != nil {
				for _, val := range r.Tracks.Tracks {
					// fmt.Printf("%+v", val)
					// break
					// fields = append(fields, slack.AttachmentField{
					// 	Title: fmt.Sprintf("%s, %s", val.Artists[0].Name, val.Album.Name),
					// 	Value: val.Name,
					// })

					attachment := slack.Attachment{
						Pretext:  fmt.Sprintf("%s, %s", val.Artists[0].Name, val.Album.Name),
						Title:    val.Name,
						ThumbURL: val.Album.Images[1].URL,
						Text:     fmt.Sprintf("@lenny add %s", getTrackId(val.Endpoint)),
						// Actions: []slack.AttachmentAction{
						// 	slack.AttachmentAction{
						// 		Name:  "AddToSpotify",
						// 		Text:  "Add",
						// 		Value: val.Endpoint,
						// 		Type:  "button",
						// 	},
						// },
						// CallbackID: "add_track",
					}
					attachments = append(attachments, attachment)
				}
			}

			if len(attachments) > 0 {
				replyChannel.Attachments = attachments
				c <- replyChannel
			}

		case "add":
			trackId := commandArray[2]

			client.AddTracksToPlaylist(userId, playlistId, spotify.ID(trackId))
			attachment := slack.Attachment{
				Pretext: "Added da track",
				Color:   "#85929E",
			}
			replyChannel.Attachments = []slack.Attachment{attachment}
			c <- replyChannel

		case "list":
			attachments := make([]slack.Attachment, 0)
			pl, _ := client.GetPlaylist(userId, playlistId)
			for _, val := range pl.Tracks.Tracks {
				attachment := slack.Attachment{
					Pretext: fmt.Sprintf("%s, %s", val.Track.Artists[0].Name, val.Track.Album.Name),
					Title:   val.Track.Name,
					Text:    fmt.Sprintf("@lenny play %s", val.Track.Endpoint),
				}
				attachments = append(attachments, attachment)
			}

			replyChannel.Attachments = attachments
			c <- replyChannel

		case "set-time":
			requestedTime := commandArray[2]

			reqTime, err := time.Parse(time.Kitchen, requestedTime)

			if err != nil {
				replyChannel.Attachments = getErrorAttachment(err, fmt.Sprintf("%s, must be in format 3:00PM/AM", err))
				c <- replyChannel
				return
			}
			t := time.Now()

			n := time.Date(t.Year(), t.Month(), t.Day(), reqTime.Hour(), reqTime.Minute(), 0, 0, t.Location())
			log.Println("RQUESTED TIME: ", n)

			standupChannel.StandupTime = n
			sc <- standupChannel

		case "set-times":
			attachments := make([]slack.Attachment, 0)
			for i := 0; i < len(setTimes); i++ {
				attachment := slack.Attachment{
					Text: fmt.Sprintf("set time: %s", setTimes[i]),
				}
				attachments = append(attachments, attachment)
			}
			replyChannel.Attachments = attachments
			c <- replyChannel
		case "set-genres":
			genresSeeds = make([]string, 0)
			if commandArray[2] == "random" {
				attachments := make([]slack.Attachment, 0)
				genres, _ := client.GetAvailableGenreSeeds()

				for i := 1; i < 5; i++ {
					rand.Seed(time.Now().UTC().UnixNano())
					genre := genres[rand.Intn(len(genres))]
					attachment := slack.Attachment{
						Text: fmt.Sprintf("added genre %s", genre),
					}
					attachments = append(attachments, attachment)

					genresSeeds = append(genresSeeds, genre)
				}

				replyChannel.Attachments = attachments
				c <- replyChannel
			} else {
				for i := 2; i < len(commandArray); i++ {
					genresSeeds = append(genresSeeds, commandArray[i])
				}
			}
		case "current":
			cp, _ := client.PlayerCurrentlyPlaying()
			attachment := slack.Attachment{
				Pretext: cp.Item.Name,
				Text:    cp.Item.Artists[0].Name,
				Color:   "#85929E",
			}
			replyChannel.Attachments = []slack.Attachment{attachment}
			c <- replyChannel
		case "genres":
			if genres, err := client.GetAvailableGenreSeeds(); err != nil {
				replyChannel.Attachments = getErrorAttachment(err, "")
				c <- replyChannel
				return
			} else {

				var genresBuffer bytes.Buffer
				for _, val := range genres {
					genresBuffer.WriteString(fmt.Sprintf("%s\n", val))
				}
				attachment := slack.Attachment{
					Text: genresBuffer.String(),
				}
				replyChannel.Attachments = []slack.Attachment{attachment}

				c <- replyChannel
			}
		case "tribute":
			track, _ := client.GetTrack("2HB4bT5bvUWDbOjNpIwmNi")
			uris := []spotify.URI{spotify.URI("spotify:track:2HB4bT5bvUWDbOjNpIwmNi")}

			opts := &spotify.PlayOptions{URIs: uris}

			client.PlayOpt(opts)

			attachment := slack.Attachment{
				Title:    "RIP Chris Cornell",
				Text:     fmt.Sprintf("%s, %s", track.Album.Name, track.Name),
				ImageURL: track.Album.Images[0].URL,
			}
			replyChannel.Attachments = []slack.Attachment{attachment}
			c <- replyChannel

		default:
			attachment := slack.Attachment{
				Pretext: "Command error",
				Text:    "I'm too smashed to play any tunes",
			}
			replyChannel.Attachments = []slack.Attachment{attachment}

			c <- replyChannel
		}
	}
}

func getErrorAttachment(err error, txt string) []slack.Attachment {
	attachment := slack.Attachment{
		Pretext: "I BROKE !!",
		Text:    fmt.Sprintf("%s", err),
	}

	if txt != "" {
		attachment.Text = txt
	}
	return []slack.Attachment{attachment}
}

func timeToPlayMusic(ch chan ReplyChannel, sc *slack.Channel) {
	var replyChannel ReplyChannel
	replyChannel.Channel = sc

	if len(genresSeeds) == 0 {
		genresSeeds = append(genresSeeds, "alternative")
	}
	seeds := spotify.Seeds{
		Genres: genresSeeds,
		//Artists: []spotify.ID{spotify.ID("5xUf6j4upBrXZPg6AI4MRK"), spotify.ID("2ziB7fzrXBoh1HUPS6sVFn")},
	}
	trackAttr := spotify.NewTrackAttributes()
	// trackAttr.MaxEnergy(1.0)
	// trackAttr.TargetEnergy(1.0)
	// trackAttr.TargetDanceability(1.0)
	opts := &spotify.Options{}

	recs, err := client.GetRecommendations(seeds, trackAttr, opts)
	// recs, err := client.Search("soundgarden", spotify.SearchTypeArtist)
	if err != nil {
		replyChannel.Attachments = getErrorAttachment(err, "")
		ch <- replyChannel
		return
	}
	log.Printf("%+v", recs)
	//playOpts := &spotify.PlayOptions{}
	uris := []spotify.URI{}
	ids := []spotify.ID{}
	for _, val := range recs.Tracks {
		trackURL := fmt.Sprintf("spotify:track:%s", val.ID)
		uris = append(uris, spotify.URI(trackURL))
		ids = append(ids, val.ID)
	}

	// for _, val := range recs.Tracks.Tracks {
	// 	trackURL := fmt.Sprintf("spotify:track:%s", val.ID)
	// 	uris = append(uris, spotify.URI(trackURL))
	// 	ids = append(ids, val.ID)
	// }

	err = client.ReplacePlaylistTracks(userId, playlistId, ids...)
	if err != nil {
		replyChannel.Attachments = getErrorAttachment(err, "")
		ch <- replyChannel
		return
	}

	//playOpts.URIs = uris
	// 	client.PlayOpt(&spotify.PlayOptions{
	// 	DeviceID: spotify.ID("11f1a54fa480f1559f1dd0cfdf42e9451dc17cb7"),
	// 	URIs:     []spotify.URI{spotify.URI("https://api.spotify.com/v1/tracks/3XVBdLihbNbxUwZosxcGuJ")},
	// })

	client.Next()
	err = client.Play()
	if err != nil {
		replyChannel.Attachments = getErrorAttachment(err, "")
		ch <- replyChannel
		return
	}

	time.Sleep(2)
	cp, _ := client.PlayerCurrentlyPlaying()

	attachment := slack.Attachment{
		Pretext: cp.Item.Name,
		Text:    cp.Item.Artists[0].Name,
		Color:   "#85929E",
	}
	replyChannel.Attachments = []slack.Attachment{attachment}
	ch <- replyChannel
}

func handleStandupTimer(c chan ReplyChannel) {
	var replyChannel ReplyChannel

	for {
		sc := <-standupChannel
		replyChannel.Channel = sc.Channel

		attachment := slack.Attachment{
			Pretext: "Set time engaged",
			Text:    fmt.Sprintf("I will blast some tunes at %s", sc.StandupTime.Format("3:04PM")),
		}

		setTime := sc.StandupTime.Format("15:04")
		newcron := gocron.NewScheduler()
		newcron.Every(1).Monday().At(setTime).Do(timeToPlayMusic, c, sc.Channel)
		newcron.Every(1).Tuesday().At(setTime).Do(timeToPlayMusic, c, sc.Channel)
		newcron.Every(1).Wednesday().At(setTime).Do(timeToPlayMusic, c, sc.Channel)
		newcron.Every(1).Thursday().At(setTime).Do(timeToPlayMusic, c, sc.Channel)
		newcron.Every(1).Friday().At(setTime).Do(timeToPlayMusic, c, sc.Channel)
		setTimes = append(setTimes, setTime)

		go func() {
			<-newcron.Start()
		}()

		replyChannel.Attachments = []slack.Attachment{attachment}
		c <- replyChannel

	}
}

func getTrackId(trackEndpoint string) string {
	strs := strings.Split(trackEndpoint, "/")
	return strs[len(strs)-1]
}

func handleReply() {
	for {
		ac := <-replyChannel
		params := slack.PostMessageParameters{}
		params.AsUser = true
		params.Attachments = ac.Attachments
		params.UnfurlMedia = true
		fmt.Println("Channel: ", ac.Channel.Name)
		_, _, errPostMessage := api.PostMessage(ac.Channel.Name, ac.DisplayTitle, params)
		if errPostMessage != nil {
			log.Fatal(errPostMessage)
		}
	}
}
