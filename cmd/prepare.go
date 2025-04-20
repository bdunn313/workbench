/*
Copyright © 2024 Brad Dunn <brad@braddunn.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/tasks/v1"
)

var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare for your upcoming week",
	Long: `Prepare for your upcoming week by reviewing calendar events, tasks,
	and answering questions about your upcoming commitments and goals.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runPrepare(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(prepareCmd)
}

type model struct {
	questions []string
	answers   []string
	index     int
	textarea  textarea.Model
	done      bool
}

func initialModel() model {
	questions := []string{
		"What shenanigans is your partner/co-parent plotting this week? 🕵️",
		"What chaos are your tiny humans orchestrating in the near future? 🎪",
		"What fresh workplace madness threatens to make your eye twitch? 🤪",
		"What social obligations have you foolishly agreed to? (No takebacks!) 🤝",
		"What important thing is your brain desperately trying to remember? 🧠",
		"Paint me a picture of your ideal Sunday victory dance! 💃",
		"Time for a brain dump! Let it all out, we won't judge. (Optional) 🗑️",
	}

	ta := textarea.New()
	ta.Placeholder = "Type your answer here..."
	ta.Focus()
	ta.CharLimit = 1000
	ta.SetWidth(80)
	ta.SetHeight(10)
	ta.ShowLineNumbers = false
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	ta.FocusedStyle.Base = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2)

	return model{
		questions: questions,
		answers:   make([]string, len(questions)),
		index:     0,
		textarea:  ta,
		done:      false,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if tea.KeyMsg(msg).Alt {
				// Alt+Enter adds a new line
				m.textarea, cmd = m.textarea.Update(msg)
				return m, cmd
			}
			if m.index < len(m.questions) {
				m.answers[m.index] = m.textarea.Value()
				m.textarea.Reset()
				m.index++
				if m.index == len(m.questions) {
					m.done = true
					return m, tea.Quit
				}
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	}

	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.done {
		return "Thank you for your answers! Analyzing your week..."
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginLeft(2)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginLeft(2)

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s\n\n%s",
		style.Render("Weekly Preparation"),
		style.Render(m.questions[m.index]),
		m.textarea.View(),
		helpStyle.Render("Press Enter to submit • Alt+Enter for new line • Ctrl+C to quit"),
	)
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[1:])
	}
	return path, nil
}

func getTokenFromFile(file string) (*oauth2.Token, error) {
	expandedPath, err := expandPath(file)
	if err != nil {
		return nil, err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(expandedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("unable to create directory: %v", err)
	}

	f, err := os.Open(expandedPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(file string, token *oauth2.Token) error {
	expandedPath, err := expandPath(file)
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(expandedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("unable to create directory: %v", err)
	}

	f, err := os.OpenFile(expandedPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
	return nil
}

func getClient(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	clientID := viper.GetString("google.client_id")
	clientSecret := viper.GetString("google.client_secret")
	tokenFile := viper.GetString("google.token_file")

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("missing Google OAuth credentials. Please set google.client_id and google.client_secret in your config file")
	}

	config.ClientID = clientID
	config.ClientSecret = clientSecret

	tok, err := getTokenFromFile(tokenFile)
	if err != nil {
		authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
		fmt.Printf("Go to the following link in your browser then type the "+
			"authorization code: \n%v\n", authURL)

		var authCode string
		if _, err := fmt.Scan(&authCode); err != nil {
			return nil, fmt.Errorf("unable to read authorization code: %v", err)
		}

		tok, err = config.Exchange(ctx, authCode)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve token from web: %v", err)
		}
		if err := saveToken(tokenFile, tok); err != nil {
			return nil, err
		}
	}
	return tok, nil
}

type openAIRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func analyzeWeek(ctx context.Context, events []*calendar.Event, taskLists []*tasks.TaskList, answers []string) (string, error) {
	apiKey := viper.GetString("openai.api_key")
	if apiKey == "" {
		return "", fmt.Errorf("missing OpenAI API key. Please set openai.api_key in your config file")
	}

	model := viper.GetString("openai.model")
	if model == "" {
		model = "gpt-4-turbo-preview" // default model
	}

	// Format calendar events
	var eventsStr strings.Builder
	for _, event := range events {
		startTime, _ := time.Parse(time.RFC3339, event.Start.DateTime)
		eventsStr.WriteString(fmt.Sprintf("- %s (%s)\n", event.Summary, startTime.Format("Mon Jan 2 15:04")))
	}

	// Format task lists
	var tasksStr strings.Builder
	for _, list := range taskLists {
		tasksStr.WriteString(fmt.Sprintf("- %s\n", list.Title))
	}

	// Format questionnaire answers
	var answersStr strings.Builder
	questions := []string{
		"Partner/co-parent updates:",
		"Kids' activities:",
		"Work commitments:",
		"Other commitments:",
		"Potential concerns:",
		"Success goals:",
		"Additional thoughts:",
	}
	for i, answer := range answers {
		if answer != "" {
			answersStr.WriteString(fmt.Sprintf("%s %s\n", questions[i], answer))
		}
	}

	prompt := fmt.Sprintf(`Analyze the following information about my upcoming week and provide a summary with insights and recommendations:

Calendar Events:
%s

Task Lists:
%s

Questionnaire Answers:
%s

Please provide:
1. A brief summary of the week's commitments
2. Potential conflicts or scheduling challenges
3. Recommendations for managing workload and stress
4. Any suggested tasks or reminders based on the information provided
5. A positive outlook or encouragement for the week ahead

Format the response in a clear, concise way with bullet points and sections.`, eventsStr.String(), tasksStr.String(), answersStr.String())

	reqBody := openAIRequest{
		Model: model,
		Messages: []message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var openAIResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return isatty.IsTerminal(f.Fd())
	}
	return false
}

func formatOutput(w io.Writer, content string) error {
	if isTerminal(w) {
		// Use Glamour for terminal output
		r, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(80),
		)
		if err != nil {
			return fmt.Errorf("error creating renderer: %v", err)
		}

		out, err := r.Render(content)
		if err != nil {
			return fmt.Errorf("error rendering markdown: %v", err)
		}

		fmt.Fprint(w, out)
	} else {
		// Plain text for non-terminal output
		fmt.Fprint(w, content)
	}
	return nil
}

func runPrepare() error {
	// Initialize Google Calendar and Tasks clients
	ctx := context.Background()
	config := &oauth2.Config{
		RedirectURL: "urn:ietf:wg:oauth:2.0:oob",
		Scopes: []string{
			calendar.CalendarReadonlyScope,
			tasks.TasksReadonlyScope,
		},
		Endpoint: google.Endpoint,
	}

	tok, err := getClient(ctx, config)
	if err != nil {
		return fmt.Errorf("unable to get client: %v", err)
	}

	client := config.Client(ctx, tok)

	// Initialize Calendar service
	calendarService, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to create Calendar service: %v", err)
	}

	// Initialize Tasks service
	tasksService, err := tasks.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to create Tasks service: %v", err)
	}

	// Get calendar events for the next week
	now := time.Now().Format(time.RFC3339)
	weekFromNow := time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339)
	events, err := calendarService.Events.List("primary").
		TimeMin(now).
		TimeMax(weekFromNow).
		SingleEvents(true).
		OrderBy("startTime").
		Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve calendar events: %v", err)
	}

	// Get tasks
	taskLists, err := tasksService.Tasklists.List().Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve task lists: %v", err)
	}

	// Run the questionnaire
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running program: %v", err)
	}

	model := m.(model)
	if !model.done {
		return fmt.Errorf("questionnaire was not completed")
	}

	// Analyze the week using OpenAI
	summary, err := analyzeWeek(ctx, events.Items, taskLists.Items, model.answers)
	if err != nil {
		return fmt.Errorf("error analyzing week: %v", err)
	}

	// Print the analysis with markdown formatting if outputting to a terminal
	return formatOutput(os.Stdout, summary)
}
