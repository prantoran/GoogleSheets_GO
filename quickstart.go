package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("sheets.googleapis.com-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	log.Printf("tokenFromFile() file: %v \n\n", file)
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)

	log.Printf("tokenFromFile()\nt.AccessToken: %v \nt.tokenType: %v\nt.RefreshToken: %v \nt.expiry: %v\n\n", t.AccessToken, t.TokenType, t.RefreshToken, t.Expiry)

	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/sheets.googleapis.com-go-quickstart.json
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets Client %v", err)
	}

	// Prints the names and majors of students in a sample spreadsheet:
	// https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
	//spreadsheetId := "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"

	// https://docs.google.com/spreadsheets/d/1zFjra05ZGfaVgKNorPdvAU-bh0QDkOn-CVoXjWtiw2w/edit
	spreadsheetId := "1zFjra05ZGfaVgKNorPdvAU-bh0QDkOn-CVoXjWtiw2w"

	//readRange := "Class Data!A2:E"
	readRange := "A3:F6"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet. %v", err)
	}

	log.Printf("\nresp: %v\n\nresp.Values: %v\n\nresp.Values type: %T \n\n", resp, resp.Values, resp.Values)

	if len(resp.Values) > 0 {
		fmt.Println("Name, Major:")
		for _, row := range resp.Values {
			// Print columns A and E, which correspond to indices 0 and 4.
			// fmt.Printf("%s, %s\n", row[0], row[4])
			fmt.Printf("Row Type: %T\n", row)
			for _, item := range row {
				fmt.Printf("%s \t", item)
			}
			fmt.Printf("\n")
		}
	} else {
		fmt.Print("No data found.")
	}

	file, err := os.Create("gsheet_result.csv")
	checkError("Cannot create file", err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, value := range resp.Values {
		cur := []string{}
		for _, item := range value {
			cur = append(cur, item.(string))
		}
		err := writer.Write(cur)
		checkError("Cannot write to file", err)
	}

}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}
