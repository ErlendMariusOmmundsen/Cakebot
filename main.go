package main

import (
	"fmt"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"log"
	"os"
	"strings"
	"time"
)

func main() {

	candidates := []string{"Hans", "Therese", "Sigurd", "Trym", "Sivert", "Asbjørn", "Erlend"}
	candidatePool := []string{"Hans", "Therese", "Sigurd", "Trym", "Sivert", "Asbjørn"}
	selectedPerson := ""
	lastDate := time.Date(2021, 10, 15, 17, 00, 00, 0, time.Local)
	timeGap, _ := time.ParseDuration("25h")

	appToken, ok := os.LookupEnv("SLACK_TOKEN")
	if !ok {
		fmt.Println("Missing SLACK_APP_TOKEN in environment")
		os.Exit(1)
	}

	botToken, ok := os.LookupEnv("SLACK_BOT_TOKEN")
	if !ok {
		fmt.Println("Missing SLACK_BOT_TOKEN in environment")
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

	go func() {
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
					if IsEnoughTimePassed(lastDate, timeGap) && ItIsMondayMyDudes() {
						lastDate = time.Now()
						chosenCandidate, newCandidates := PopCandidate(candidatePool)
						if len(newCandidates) == 0 {
							candidatePool = candidates
						} else {
							candidatePool = newCandidates
						}
						for i := 0; i < len(candidatePool); i++ {
							println(candidatePool[i])
						}
						selectedPerson = chosenCandidate
						chosenMsg := "Gratulerer, " + chosenCandidate + ". Det er din tur til å lage kake! :cake:"
						switch ev := innerEvent.Data.(type) {
						case *slackevents.AppMentionEvent:
							msg := slack.Attachment{
								Title:     "Artig link",
								TitleLink: "https://youtu.be/WJq4jWSQNd8",
								Pretext:   chosenMsg,
								Text:      "Du kan jo prøve deg på denne:",
								ImageURL:  "https://www.boredpanda.com/blog/wp-content/uploads/2020/10/funny-expectation-reality-cakes-14-5f7f16831f8db__700.jpg",
							}
							_, _, err := api.PostMessage(ev.Channel, slack.MsgOptionAttachments(msg))
							if err != nil {
								fmt.Printf("failed posting message: %v", err)
							}
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

				payload := map[string]interface{}{
					"blocks": []slack.Block{
						slack.NewSectionBlock(
							&slack.TextBlockObject{
								Type: slack.MarkdownType,
								Text: "foo",
							},
							nil,
							slack.NewAccessory(
								slack.NewButtonBlockElement(
									"",
									"some value",
									&slack.TextBlockObject{
										Type: slack.PlainTextType,
										Text: "bar",
									},
								),
							),
						),
					}}

				client.Ack(*evt.Request, payload)
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
