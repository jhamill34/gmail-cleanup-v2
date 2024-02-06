package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.MailGoogleComScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}
	user := "me"

	senders := make(map[string][]string)

	var token string
	for {
		messagesReq := srv.Users.Messages.
			List(user).
			Q("in:inbox").
			MaxResults(500)

		if token != "" {
			messagesReq = messagesReq.PageToken(token)
		}

		messages, err := messagesReq.Do()
		if err != nil {
			log.Fatalf("Unable to retrieve page of messages: %v", err)
		}

		fmt.Printf("Found %d messages\n", len(messages.Messages))

		token = messages.NextPageToken
		for _, message := range messages.Messages {
			msg, err := srv.Users.Messages.Get(user, message.Id).Format("metadata").Do()
			if err != nil {
				log.Fatalf("Unable to retrieve message %v: %v", message.Id, err)
			}

			for _, header := range msg.Payload.Headers {
				if header.Name == "From" {
					addr, err := mail.ParseAddress(header.Value)
					if err != nil {
						log.Printf("Unable to parse mail address: %v\n", err)
						continue
					}

					if _, ok := senders[addr.Address]; !ok {
						senders[addr.Address] = []string{}
					}

					senders[addr.Address] = append(senders[addr.Address], msg.Id)
				}
			}
		}

		if token == "" {
			break
		}
	}

	sendersFile, err := os.OpenFile("senders.json", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Unable to open senders file: %v", err)
	}

	err = json.NewEncoder(sendersFile).Encode(senders)
	if err != nil {
		log.Fatalf("Unable to encode senders: %v", err)
	}

}
