package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/net/idna"
)

type Response struct {
	Code  int    `json:"code"`
	State string `json:"state"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
	Results Results `json:"results"`
}

type Results struct {
	Domain     string `json:"domain"`
	Servername string `json:"servername"`
	Tld        string `json:"tld"`
	Registered bool   `json:"registered"`
	Reserved   bool   `json:"reserved"`
	ClientHold bool   `json:"client_hold"`
	Detail     struct {
		Registrant []string `json:"registrant"`
		Admin      []string `json:"admin"`
		Tech       []string `json:"tech"`
		Billing    []string `json:"billing"`
		Status     []string `json:"status"`
		Date       []string `json:"date"`
		NameServer []string `json:"name_server"`
	} `json:"detail"`
	Raw []string `json:"raw"`
}

func whois(session *discordgo.Session, message *discordgo.MessageCreate, domains []string) {
	p := idna.New()

	for _, s := range domains {
		domain, err := p.ToASCII(s)
		if err != nil {
			log.Println(err)
			continue
		}

		res := request(domain)
		if res.Code != 200 {
			continue
		}

		embed := createEmbed(s, res, session)

		session.ChannelMessageSendEmbed(message.ChannelID, embed)
	}
}

func request(domain string) (response Response) {
	url := "https://api.whoisproxy.info/whois/" + domain

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println(err.Error())
	}

	client := new(http.Client)
	res, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err.Error())
	}

	response = Response{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println(err.Error())
	}

	return
}

func createEmbed(domain string, res Response, session *discordgo.Session) (embed *discordgo.MessageEmbed) {
	registrant := strings.Join(res.Results.Detail.Registrant, "\n")
	date := strings.Join(res.Results.Detail.Date, "\n")
	nameServers := strings.Join(res.Results.Detail.NameServer, "\n")

	// ??????????????????????????????????????????????????????????????????
	if len(res.Results.Detail.Registrant) == 0 {
		registrant = "????????????"
	}
	if len(res.Results.Detail.Date) == 0 {
		date = "????????????"
	}
	if len(res.Results.Detail.NameServer) == 0 {
		nameServers = "????????????"
	}

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "????????????",
			Value:  "",
			Inline: false,
		},
		{
			Name:   "?????????",
			Value:  registrant,
			Inline: false,
		},
		{
			Name:   "??????",
			Value:  date,
			Inline: false,
		},
		{
			Name:   "?????????????????????",
			Value:  nameServers,
			Inline: false,
		},
	}

	if res.Results.Registered {
		fields[0].Value = "????????????"
	} else {
		fields[0].Value = "?????????"
		fields = fields[:1]
	}

	embed = &discordgo.MessageEmbed{
		Title: domain,
		URL:   "http://" + domain,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    session.State.User.Username,
			IconURL: session.State.User.AvatarURL(""),
		},
		Color:  0x00bfff,
		Fields: fields,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "https://whoisproxy.info ??????",
		},
	}

	return
}
