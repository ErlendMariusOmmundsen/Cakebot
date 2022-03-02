package main

import (
	"fmt"
	"github.com/peterhellberg/giphy"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"log"
	"os"
	"strings"
	"time"
)

func main() {

	candidates := []string{"Hans", "Therese", "Trym", "Sivert", "Asbjørn", "Erlend"}
	candidatePool := []string{"Hans", "Therese", "Trym", "Sivert", "Asbjørn", "Erlend"}
	selectedPerson := ""
	lastDate := time.Date(2021, 10, 15, 17, 00, 00, 0, time.Local)
	cooldown, _ := time.ParseDuration("10m")
	searchTerms := []string{"cake"}

	appToken, ok := os.LookupEnv("SLACK_APP_TOKEN")
	if !ok {
		fmt.Println("Missing SLACK_APP_TOKEN in environment")
		os.Exit(1)
	}

	botToken, ok := os.LookupEnv("SLACK_BOT_TOKEN")
	if !ok {
		fmt.Println("Missing SLACK_BOT_TOKEN in environment")
		os.Exit(1)
	}

	giphyKey, ok := os.LookupEnv("GIPHY_API_KEY")
	if !ok {
		fmt.Println("Missing GIPHY_API_KEY in environment")
		os.Exit(1)
	}

	if !strings.HasPrefix(appToken, "xapp-") {
		_, err := fmt.Fprintf(os.Stderr, "SLACK_APP_TOKEN must have the prefix \"xapp-\".\n")
		if err != nil {
			log.Println(err)
		}
	}

	if !strings.HasPrefix(botToken, "xoxb-") {
		_, err := fmt.Fprintf(os.Stderr, "SLACK_BOT_TOKEN must have the prefix \"xoxb-\".\n")
		if err != nil {
			log.Println(err)
		}
	}

	api := slack.New(
		botToken,
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
		slack.OptionAppLevelToken(appToken))

	client := socketmode.New(
		api,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	// TODO: Add funtionality for manually pick next candidate
	go func() {

		giphyClient := giphy.NewClient(giphy.APIKey(giphyKey))
		for evt := range client.Events {
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				fmt.Println("Connecting to Slack with Socket Mode...")
			case socketmode.EventTypeConnectionError:
				fmt.Println("Connection failed. Retrying later...")
			case socketmode.EventTypeConnected:
				fmt.Println("Connected to Slack with Socket Mode.")
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					fmt.Printf("Ignored %+v\n", evt)

					continue
				}

				fmt.Printf("Event received: %+v\n", eventsAPIEvent)

				client.Ack(*evt.Request)

				switch eventsAPIEvent.Type {
				case slackevents.CallbackEvent:
					innerEvent := eventsAPIEvent.InnerEvent
					if IsEnoughTimePassed(lastDate, cooldown) {
						lastDate = time.Now()
						chosenCandidate, newCandidates := PopCandidate(candidatePool)
						if len(newCandidates) == 0 {
							candidatePool = resetCandidates(candidatePool, candidates)
						} else {
							candidatePool = newCandidates
						}
						for i := 0; i < len(candidatePool); i++ {
							println(candidatePool[i])
						}
						selectedPerson = chosenCandidate
						chosenMsg := "<!channel> Gratulerer, " + chosenCandidate + ". Det er din tur til å lage kake! :cake:"
						switch event := innerEvent.Data.(type) {
						case *slackevents.AppMentionEvent:
							msg := slack.Attachment{
								Title:     "Artig link",
								TitleLink: "https://youtu.be/WJq4jWSQNd8",
								Pretext:   chosenMsg,
								Text:      "Bli inspirert:",
							}
							random, err := giphyClient.Random(searchTerms)
							if err != nil {
								fmt.Println(err.Error())
								msg.ImageURL = "https://img.devrant.com/devrant/rant/r_2306733_137EK.jpg"
							} else {
								msg.ImageURL = random.Data.MediaURL()
							}
							_, respTS, err := api.PostMessage(event.Channel, slack.MsgOptionAttachments(msg))
							if err != nil {
								fmt.Printf("failed posting message: %v", err)
							}
							pinErr := api.AddPin(event.Channel, slack.NewRefToMessage(event.Channel, respTS))
							if pinErr != nil {
								fmt.Printf("Error adding pin: %s\n", err)
							}
							// TODO: Remove existing pins
						}
					} else {
						log.Println("Cakebot was called, but not enough time has passed")
						switch ev := innerEvent.Data.(type) {
						case *slackevents.AppMentionEvent:
							msg := slack.Attachment{
								Pretext: "Det er fortsatt " + selectedPerson + " som skal lage kake :cake:",
							}
							_, _, err := api.PostMessage(ev.Channel, slack.MsgOptionAttachments(msg))
							if err != nil {
								fmt.Printf("failed posting message: %v", err)
							}
						}
					}

				default:
					client.Debugf("unsupported Events API event received")
				}
			case socketmode.EventTypeInteractive:
				callback, ok := evt.Data.(slack.InteractionCallback)
				if !ok {
					fmt.Printf("Ignored %+v\n", evt)

					continue
				}

				fmt.Printf("Interaction received: %+v\n", callback)

				var payload interface{}

				switch callback.Type {
				case slack.InteractionTypeBlockActions:
					// See https://api.slack.com/apis/connections/socket-implement#button

					client.Debugf("button clicked!")
				case slack.InteractionTypeShortcut:
				case slack.InteractionTypeViewSubmission:
					// See https://api.slack.com/apis/connections/socket-implement#modal
				case slack.InteractionTypeDialogSubmission:
				default:

				}

				client.Ack(*evt.Request, payload)
			case socketmode.EventTypeSlashCommand:
				cmd, ok := evt.Data.(slack.SlashCommand)
				if !ok {
					fmt.Printf("Ignored %+v\n", evt)

					continue
				}

				client.Debugf("Slash command received: %+v", cmd)

				profile, _ := api.GetUserProfile(&slack.GetUserProfileParameters{UserID: cmd.UserID})
				switch {
				case cmd.Command == "/kandidater":
					msg := slack.Attachment{
						Pretext: profile.FirstName + " kalte på meg :kanelsnurr: Dette er kandidatene:",
						Text:    GetStringsOfSlice(candidatePool),
					}
					_, _, err := api.PostMessage(cmd.ChannelID, slack.MsgOptionAttachments(msg))
					if err != nil {
						fmt.Printf("failed posting message: %v", err)
					} else {
						fmt.Printf("Kandidater: %v\n", candidatePool)
					}

				case cmd.Command == "/reset":
					profile, _ := api.GetUserProfile(&slack.GetUserProfileParameters{UserID: cmd.UserID})
					candidatePool = resetCandidates(candidatePool, candidates)
					lastDate = time.Date(2000, 0, 0, 0, 0, 0, 0, time.Local)
					msg := slack.Attachment{
						Pretext: profile.FirstName + " la alle til i trekningen! :powerstonk: Dette er nå kandidatene: ",
						Text:    GetStringsOfSlice(candidatePool),
					}
					_, _, err := api.PostMessage(cmd.ChannelID, slack.MsgOptionAttachments(msg))
					if err != nil {
						fmt.Printf("failed posting message: %v", err)
					}

				case cmd.Command == "/fjern_kandidat":
					profile, _ := api.GetUserProfile(&slack.GetUserProfileParameters{UserID: cmd.UserID})
					var msg slack.Attachment
					if Contains(candidatePool, cmd.Text) {
						msg = slack.Attachment{
							Pretext: profile.FirstName + " fjernet " + cmd.Text + " fra trekningen :notstonks:",
						}
						candidatePool = Remove(candidatePool, GetIndexInSlice(candidatePool, cmd.Text))
					} else {
						msg = slack.Attachment{
							Pretext: profile.FirstName + " prøvde å fjerne " + cmd.Text + " fra trekningen, men " + cmd.Text + " var aldri med i trekningen :shrek:",
						}
					}
					_, _, err := api.PostMessage(cmd.ChannelID, slack.MsgOptionAttachments(msg))
					if err != nil {
						fmt.Printf("failed posting message: %v", err)
					}

				case cmd.Command == "/legg_til_kandidat":
					profile, _ := api.GetUserProfile(&slack.GetUserProfileParameters{UserID: cmd.UserID})
					var msg slack.Attachment
					if !Contains(candidatePool, cmd.Text) && len(cmd.Text) > 2 {
						candidatePool = append(candidatePool, cmd.Text)
						msg = slack.Attachment{
							Pretext: profile.FirstName + " la til " + cmd.Text + " i trekningen :powerstonk:",
						}
					} else {
						msg = slack.Attachment{
							Pretext: profile.FirstName + " prøvde å legge til \"" + cmd.Text + "\" i trekningen, men \"" + cmd.Text + "\" er for kort eller allerede med i trekningen :shrek:",
						}
					}
					_, _, err := api.PostMessage(cmd.ChannelID, slack.MsgOptionAttachments(msg))
					if err != nil {
						fmt.Printf("failed posting message: %v", err)
					}
				case cmd.Command == "/velg_kandidat":
					profile, _ := api.GetUserProfile(&slack.GetUserProfileParameters{UserID: cmd.UserID})
					var msg slack.Attachment
					chosenMsg := "<!channel> Gratulerer, " + cmd.Text + " :cake: " + profile.FirstName + " valgte deg, det er din tur til å lage kake! :cake:"
					msg = slack.Attachment{
						Title:     "Artig link",
						TitleLink: "https://youtu.be/WJq4jWSQNd8",
						Pretext:   chosenMsg,
						Text:      "Bli inspirert:",
					}
					random, err := giphyClient.Random(searchTerms)
					if err != nil {
						fmt.Println(err.Error())
						msg.ImageURL = "https://img.devrant.com/devrant/rant/r_2306733_137EK.jpg"
					} else {
						msg.ImageURL = random.Data.MediaURL()
					}
					_, respTS, err := api.PostMessage(cmd.ChannelID, slack.MsgOptionAttachments(msg))
					if err != nil {
						fmt.Printf("failed posting message: %v", err)
					}
					pinErr := api.AddPin(cmd.ChannelID, slack.NewRefToMessage(cmd.ChannelID, respTS))
					if pinErr != nil {
						fmt.Printf("Error adding pin: %s\n", err)
					}
					// TODO: Remove existing pins
					if Contains(candidatePool, cmd.Text) && len(cmd.Text) > 2 {
						candidatePool = Remove(candidatePool, GetIndexInSlice(candidatePool, cmd.Text))
						if len(candidatePool) == 0 {
							candidatePool = resetCandidates(candidatePool, candidates)
						}
					}
					_, _, err = api.PostMessage(cmd.ChannelID, slack.MsgOptionAttachments(msg))
					if err != nil {
						fmt.Printf("failed posting message: %v", err)
					}
				}
			default:
				_, err := fmt.Fprintf(os.Stderr, "Unexpected event type received: %s\n", evt.Type)
				if err != nil {
					err.Error()
				}
			}
		}
	}()

	err := client.Run()
	if err != nil {
		err.Error()
	}

}
