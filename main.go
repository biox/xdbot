package main // import "github.com/biox/xdbot"

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/thoj/go-ircevent"
)

var xdClient = &http.Client{Timeout: 10 * time.Second}
var ccCommand = regexp.MustCompile(`!cc \w+`)

type cryptoComparePayload struct {
	USD float64 `json:"USD"`
}

type coinbaseBTCPayload struct {
	Time struct {
		Updated    string    `json:"updated"`
		UpdatedISO time.Time `json:"updatedISO"`
		Updateduk  string    `json:"updateduk"`
	} `json:"time"`
	Disclaimer string `json:"disclaimer"`
	Bpi        struct {
		USD struct {
			Code        string  `json:"code"`
			Rate        string  `json:"rate"`
			Description string  `json:"description"`
			RateFloat   float64 `json:"rate_float"`
		} `json:"USD"`
	} `json:"bpi"`
}

const (
	roomName              = "#xddd"
	coinbaseBTCEndpoint   = "https://api.coindesk.com/v1/bpi/currentprice/USD.json"
	cryptoCompareEndpoint = "https://min-api.cryptocompare.com/data/price?fsym=SYMBOL&tsyms=USD"
)

func main() {
	con := irc.IRC("xdbot", "xdbot")
	err := con.Connect("irc.freenode.net:6667")

	if err != nil {
		fmt.Println("Failed to connect")
		return
	}

	con.AddCallback("001", func(e *irc.Event) {
		con.Join(roomName)
	})

	// Static responses
	con.AddCallback("PRIVMSG", func(e *irc.Event) {
		switch e.Message() {
		case "what is the airspeed velocity of an unladen swallow?":
			con.Privmsg(roomName, "african or european?")
		case "i hate you":
			con.Privmsg(roomName, "same")
		case "!w":
			response, err := getWeather()
			if err != nil {
				con.Privmsg(roomName, "failed to fetch weather because biox is a shit programmer")
				break
			}
			con.Privmsg(roomName, response)
		case "!btc":
			response, err := getBTC()
			if err != nil {
				con.Privmsg(roomName, "failed to fetch btc because biox is a shit programmer")
				break
			}
			con.Privmsg(roomName, response)
		}

		// More advanced logicks
		switch {
		case ccCommand.MatchString(e.Message()):
			coin := strings.Split(e.Message(), "!cc ")
			coinToQuery := strings.ToUpper(coin[1])
			response, err := getCrypto(coinToQuery)
			if err != nil {
				con.Privmsg(roomName, "failed to fetch the price of "+coinToQuery)
				break
			}
			con.Privmsg(roomName, response)
		}
	})

	// Make a new callback that accepts !cc (symbol) and passes symbol through to the getCC function

	con.Loop()
}

func getWeather() (string, error) {
	return "shitty", errors.New("this text literally doesnt matter")
}

func getBTC() (string, error) {

	r, err := xdClient.Get(coinbaseBTCEndpoint)
	if err != nil {
		return "", err
	}

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	btcBody := coinbaseBTCPayload{}

	err = json.Unmarshal(body, &btcBody)
	if err != nil {
		return "", err
	}

	response := strings.Split(btcBody.Bpi.USD.Rate, ".")

	return "$" + response[0], err
}

func getCrypto(symbol string) (string, error) {
	upperCaseSymbol := strings.ToUpper(symbol)

	url := strings.Replace(cryptoCompareEndpoint, "SYMBOL", upperCaseSymbol, -1)

	r, err := xdClient.Get(url)
	if err != nil {
		return "", err
	}

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	coinBody := cryptoComparePayload{}

	err = json.Unmarshal(body, &coinBody)
	if err != nil {
		return "", err
	}

	return "$" + strconv.FormatFloat(coinBody.USD, 'f', 2, 64), err
}
