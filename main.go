package main

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"net/http"
	"strings"
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
	//p := tea.NewProgram(initialModel())
	//if _, err := p.Run(); err != nil {
	//	fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	//}

	fmt.Println(getWeatherForecast("Raipur"))

}

func getWeatherForecast(cityName string) string {
	apikey := "provided in .env"
	urlLocation := "http://api.openweathermap.org/geo/1.0/direct?q=" + cityName + "&limit=1&appid=" + apikey

	res, err := http.Get(urlLocation)
	if err != nil {
		return "Could not fetch Weather Forecast"
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)

	var locations []Location
	err = json.Unmarshal(body, &locations)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return "Error decoding json"
	}
	lat := locations[0].Lat
	lon := locations[0].Lon

	//for _, location := range locations {
	//	fmt.Printf("Name: %s, Lat: %.6f, Lon: %.6f\n", location.Name, location.Lat, location.Lon)
	//}

	urlWeather := "https://api.openweathermap.org/data/2.5/weather?lat=" +
		fmt.Sprintf("%.5f", lat) + "&lon=" +
		fmt.Sprintf("%.5f", lon) + "&appid=" + apikey

	res, err = http.Get(urlWeather)
	if err != nil {
		//fmt.Println("Error fetching URL:", err)
		return "Could not fetch Weather Forecast"
	}
	defer res.Body.Close()

	body, err = io.ReadAll(res.Body)
	if err != nil {
		//fmt.Println("Error reading response body:", err)
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

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
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

			if v == "" {
				// Don't send empty messages.
				return m, nil
			}

			weatherForecast := getWeatherForecast(m.cityName[0])

			m.cityName = append(m.cityName, m.senderStyle.Render("You: ")+v)
			m.viewport.SetContent(strings.Join(m.cityName, "\n") +
				weatherForecast)
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
