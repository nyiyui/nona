package nona

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Webhook struct {
	URL   string             `json:"url"`
	ID    string             `json:"id"`
	Token string             `json:"token"`
	s     *discordgo.Session `json:"-"`
	cl    *http.Client       `json:"-"`
}

var _ Log = (*Webhook)(nil)

func NewWebhook(config json.RawMessage) (Log, error) {
	w := Webhook{
		cl: &http.Client{},
	}
	err := json.Unmarshal(config, &w)
	if err != nil {
		return nil, err
	}
	if w.URL != "" {
		u, err := url.Parse(w.URL)
		if err != nil {
			return nil, err
		}
		splitted := strings.Split(u.Path, "/")
		if len(splitted) != 5 {
			return nil, fmt.Errorf("invalid webhook url: %s", w.URL)
		}
		w.ID = splitted[2]
		w.Token = splitted[3]
	}
	w.s, err = discordgo.New("")
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (w *Webhook) Handle(key string, r *http.Request) error {
	log.Println("webhook", key, r.Method, r.URL.Path)
	header := new(bytes.Buffer)
	err := r.Header.Write(header)
	if err != nil {
		return err
	}
	remote := r.RemoteAddr
	if remote2 := r.Header.Get("X-Forwarded-For"); remote2 != "" {
		remote = remote2
	}
	body := &discordgo.WebhookParams{
		Content: fmt.Sprintf("From %s for %s", remote, key),
		Embeds: []*discordgo.MessageEmbed{
			{
				Title: fmt.Sprintf("Request from %s", remote),
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Meta",
						Value:  fmt.Sprintf("%s %s %s", r.Proto, r.Method, r.URL.Path),
						Inline: true,
					},
					{
						Name:   "Remote",
						Value:  remote,
						Inline: true,
					},
					{
						Name:   "Headers",
						Value:  "```\n" + header.String() + "\n```",
						Inline: false,
					},
				},
			},
		},
	}
	body2 := new(bytes.Buffer)
	err = json.NewEncoder(body2).Encode(body)
	if err != nil {
		return err
	}
	log.Println("webhook", key, r.Method, r.URL.Path)
	resp, err := w.cl.Post(w.URL, "application/json", body2)
	if err != nil {
		return err
	}
	log.Print("resp")
	log.Print(resp)
	return nil
}
