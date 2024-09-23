package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	viewport    viewport.Model
	cityName    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
}

type Location struct {
	Name       string            `json:"name"`
	LocalNames map[string]string `json:"local_names"`
	Lat        float64           `json:"lat"`
	Lon        float64           `json:"lon"`
	Country    string            `json:"country"`
	State      string            `json:"state"`
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	}
}

func getWeatherForecast(cityName string) string {
	apikey := "provided in .env"
	urlLocation := "http://api.openweathermap.org/geo/1.0/direct?q=" + cityName + "&limit=1&appid=" + apikey

	res, err := http.Get(urlLocation)
	if err != nil {
		return "Could not fetch Weather Forecast"
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		return fmt.Sprintf("Error: received status code %d", res.StatusCode)
	}

	var locations []Location
	err = json.Unmarshal(body, &locations)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return "Error decoding json"
	}
	lat := locations[0].Lat
	lon := locations[0].Lon

	urlWeather := "https://api.openweathermap.org/data/2.5/weather?lat=" +
		fmt.Sprintf("%.5f", lat) + "&lon=" +
		fmt.Sprintf("%.5f", lon) + "&appid=" + apikey

	res, err = http.Get(urlWeather)
	if err != nil {
		return "Could not fetch Weather Forecast"
	}
	defer res.Body.Close()

	body, err = io.ReadAll(res.Body)
	if err != nil {
		return "Error reading response body:"
	}

	return string(body)
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Enter a city name..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 30

	ta.SetWidth(30)
	ta.SetHeight(1)

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(20, 10)
	vp.SetContent(`Welcome to the weather forecast!`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:    ta,
		cityName:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {

		case "esc", "ctrl+c":
			fmt.Println(m.textarea.Value())
			return m, tea.Quit

		case "enter":
			v := m.textarea.Value()

			weatherForecast := getWeatherForecast(v)
			formattedData := fmt.Sprintf("\n%s", weatherForecast)

			m.cityName = append(m.cityName, m.senderStyle.Render("You: ")+v+formattedData)

			m.viewport.SetContent(strings.Join(m.cityName, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()
			return m, nil

		default:
			// Send all other key presses to the textarea.
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}

	case cursor.BlinkMsg:
		// Textarea should also process cursor blinks.
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd

	default:
		return m, nil
	}
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}
