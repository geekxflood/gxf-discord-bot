package handlers

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/alitto/pond/v2"
	"github.com/bwmarrin/discordgo"
	"github.com/geekxflood/common/logging"
	"github.com/geekxflood/gxf-discord-bot/internal/auth"
	"github.com/geekxflood/gxf-discord-bot/internal/config"
	"github.com/robfig/cron/v3"
)

// ActionManager manages all bot actions
type ActionManager struct {
	actions     []*Action
	cfg         config.Provider
	authMgr     *auth.Manager
	logger      logging.Logger
	rateLimiter *RateLimiter
	scheduler   *cron.Cron
	schedulerMu sync.Mutex
	workerPool  pond.Pool
}

// Action represents a configured bot action
type Action struct {
	Name        string
	Description string
	Type        string
	Trigger     *Trigger
	Response    *Response
	Conditions  []*Condition
	RateLimit   *RateLimitConfig
	RequireAuth bool
	Pattern     *regexp.Regexp // Compiled regex for message type
}

// Trigger defines what triggers an action
type Trigger struct {
	Command  string
	Pattern  string
	Emoji    string
	Schedule string
	Channels []string
	Guilds   []string
}

// Response defines how the bot responds
type Response struct {
	Type       string
	Content    string
	Embed      *Embed
	Reaction   string
	HTTP       *HTTPConfig
	WebhookURL string
	DeleteAfter int
	Ephemeral  bool
}

// Embed represents a Discord embed
type Embed struct {
	Title       string
	Description string
	Color       int
	Fields      []EmbedField
	Footer      string
	Image       string
	Thumbnail   string
	Author      *EmbedAuthor
	Timestamp   bool
}

// EmbedField represents a field in an embed
type EmbedField struct {
	Name   string
	Value  string
	Inline bool
}

// EmbedAuthor represents embed author
type EmbedAuthor struct {
	Name    string
	IconURL string
	URL     string
}

// HTTPConfig for HTTP response type
type HTTPConfig struct {
	URL     string
	Method  string
	Headers map[string]string
	Body    string
	Timeout int
}

// Condition defines execution conditions
type Condition struct {
	Type     string
	Value    string
	Operator string
}

// RateLimitConfig defines rate limiting
type RateLimitConfig struct {
	Requests int
	Window   int
	Scope    string
}

// NewActionManager creates a new action manager
func NewActionManager(cfg config.Provider, authMgr *auth.Manager, logger logging.Logger) (*ActionManager, error) {
	// Create worker pool for concurrent action execution
	// Pool with 10 workers, max 100 tasks in queue
	pool := pond.NewPool(10, pond.WithMaxCapacity(100))

	mgr := &ActionManager{
		cfg:         cfg,
		authMgr:     authMgr,
		logger:      logger,
		rateLimiter: NewRateLimiter(),
		scheduler:   cron.New(cron.WithSeconds()),
		workerPool:  pool,
	}

	// Load actions from configuration
	if err := mgr.loadActions(); err != nil {
		return nil, fmt.Errorf("failed to load actions: %w", err)
	}

	logger.Info("Action manager initialized", "actions", len(mgr.actions), "workers", pool.MaxWorkers())
	return mgr, nil
}

// loadActions loads actions from configuration
func (m *ActionManager) loadActions() error {
	m.actions = make([]*Action, 0)

	// Iterate through actions array
	for i := 0; ; i++ {
		key := fmt.Sprintf("actions[%d]", i)
		if !m.cfg.Exists(key) {
			break
		}

		action, err := m.loadAction(i)
		if err != nil {
			return fmt.Errorf("failed to load action at index %d: %w", i, err)
		}

		m.actions = append(m.actions, action)
	}

	return nil
}

// loadAction loads a single action from configuration
func (m *ActionManager) loadAction(index int) (*Action, error) {
	prefix := fmt.Sprintf("actions[%d]", index)

	name, _ := m.cfg.GetString(prefix+".name", "")
	description, _ := m.cfg.GetString(prefix+".description", "")
	actionType, _ := m.cfg.GetString(prefix+".type", "")

	action := &Action{
		Name:        name,
		Description: description,
		Type:        actionType,
	}

	// Load trigger
	trigger := &Trigger{}
	if m.cfg.Exists(prefix + ".trigger.command") {
		trigger.Command, _ = m.cfg.GetString(prefix+".trigger.command", "")
	}
	if m.cfg.Exists(prefix + ".trigger.pattern") {
		trigger.Pattern, _ = m.cfg.GetString(prefix+".trigger.pattern", "")
		// Compile regex
		pattern, err := regexp.Compile(trigger.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
		action.Pattern = pattern
	}
	if m.cfg.Exists(prefix + ".trigger.emoji") {
		trigger.Emoji, _ = m.cfg.GetString(prefix+".trigger.emoji", "")
	}
	if m.cfg.Exists(prefix + ".trigger.schedule") {
		trigger.Schedule, _ = m.cfg.GetString(prefix+".trigger.schedule", "")
	}
	if m.cfg.Exists(prefix + ".trigger.channels") {
		trigger.Channels, _ = m.cfg.GetStringSlice(prefix + ".trigger.channels")
	}
	if m.cfg.Exists(prefix + ".trigger.guilds") {
		trigger.Guilds, _ = m.cfg.GetStringSlice(prefix + ".trigger.guilds")
	}
	action.Trigger = trigger

	// Load response
	response, err := m.loadResponse(prefix + ".response")
	if err != nil {
		return nil, fmt.Errorf("failed to load response: %w", err)
	}
	action.Response = response

	// Load conditions
	action.Conditions = m.loadConditions(prefix)

	// Load rate limit
	if m.cfg.Exists(prefix + ".rateLimit") {
		action.RateLimit = m.loadRateLimit(prefix + ".rateLimit")
	}

	// Load auth requirement
	if m.cfg.Exists(prefix + ".requireAuth") {
		action.RequireAuth, _ = m.cfg.GetBool(prefix+".requireAuth", false)
	}

	return action, nil
}

// loadResponse loads response configuration
func (m *ActionManager) loadResponse(prefix string) (*Response, error) {
	responseType, _ := m.cfg.GetString(prefix+".type", "")

	response := &Response{
		Type: responseType,
	}

	if m.cfg.Exists(prefix + ".content") {
		response.Content, _ = m.cfg.GetString(prefix+".content", "")
	}

	if m.cfg.Exists(prefix + ".reaction") {
		response.Reaction, _ = m.cfg.GetString(prefix+".reaction", "")
	}

	if m.cfg.Exists(prefix + ".webhookUrl") {
		response.WebhookURL, _ = m.cfg.GetString(prefix+".webhookUrl", "")
	}

	if m.cfg.Exists(prefix + ".deleteAfter") {
		response.DeleteAfter, _ = m.cfg.GetInt(prefix+".deleteAfter", 0)
	}

	if m.cfg.Exists(prefix + ".ephemeral") {
		response.Ephemeral, _ = m.cfg.GetBool(prefix+".ephemeral", false)
	}

	// Load embed
	if m.cfg.Exists(prefix + ".embed") {
		response.Embed = m.loadEmbed(prefix + ".embed")
	}

	// Load HTTP config
	if m.cfg.Exists(prefix + ".http") {
		response.HTTP = m.loadHTTPConfig(prefix + ".http")
	}

	return response, nil
}

// loadEmbed loads embed configuration
func (m *ActionManager) loadEmbed(prefix string) *Embed {
	embed := &Embed{}

	if m.cfg.Exists(prefix + ".title") {
		embed.Title, _ = m.cfg.GetString(prefix+".title", "")
	}
	if m.cfg.Exists(prefix + ".description") {
		embed.Description, _ = m.cfg.GetString(prefix+".description", "")
	}
	if m.cfg.Exists(prefix + ".color") {
		embed.Color, _ = m.cfg.GetInt(prefix+".color", 0)
	}
	if m.cfg.Exists(prefix + ".footer") {
		embed.Footer, _ = m.cfg.GetString(prefix+".footer", "")
	}
	if m.cfg.Exists(prefix + ".image") {
		embed.Image, _ = m.cfg.GetString(prefix+".image", "")
	}
	if m.cfg.Exists(prefix + ".thumbnail") {
		embed.Thumbnail, _ = m.cfg.GetString(prefix+".thumbnail", "")
	}
	if m.cfg.Exists(prefix + ".timestamp") {
		embed.Timestamp, _ = m.cfg.GetBool(prefix+".timestamp", false)
	}

	// Load fields
	for i := 0; ; i++ {
		fieldPrefix := fmt.Sprintf("%s.fields[%d]", prefix, i)
		if !m.cfg.Exists(fieldPrefix) {
			break
		}

		name, _ := m.cfg.GetString(fieldPrefix+".name", "")
		value, _ := m.cfg.GetString(fieldPrefix+".value", "")
		inline, _ := m.cfg.GetBool(fieldPrefix+".inline", false)

		embed.Fields = append(embed.Fields, EmbedField{
			Name:   name,
			Value:  value,
			Inline: inline,
		})
	}

	// Load author
	if m.cfg.Exists(prefix + ".author") {
		name, _ := m.cfg.GetString(prefix+".author.name", "")
		iconURL, _ := m.cfg.GetString(prefix+".author.iconUrl", "")
		url, _ := m.cfg.GetString(prefix+".author.url", "")
		embed.Author = &EmbedAuthor{
			Name:    name,
			IconURL: iconURL,
			URL:     url,
		}
	}

	return embed
}

// loadHTTPConfig loads HTTP configuration
func (m *ActionManager) loadHTTPConfig(prefix string) *HTTPConfig {
	cfg := &HTTPConfig{}

	cfg.URL, _ = m.cfg.GetString(prefix+".url", "")
	cfg.Method, _ = m.cfg.GetString(prefix+".method", "POST")
	cfg.Body, _ = m.cfg.GetString(prefix+".body", "")
	cfg.Timeout, _ = m.cfg.GetInt(prefix+".timeout", 30)

	// Load headers
	if m.cfg.Exists(prefix + ".headers") {
		headersMap, _ := m.cfg.GetMap(prefix + ".headers")
		cfg.Headers = make(map[string]string)
		for k, v := range headersMap {
			if str, ok := v.(string); ok {
				cfg.Headers[k] = str
			}
		}
	}

	return cfg
}

// loadConditions loads conditions
func (m *ActionManager) loadConditions(prefix string) []*Condition {
	conditions := make([]*Condition, 0)

	for i := 0; ; i++ {
		condPrefix := fmt.Sprintf("%s.conditions[%d]", prefix, i)
		if !m.cfg.Exists(condPrefix) {
			break
		}

		condType, _ := m.cfg.GetString(condPrefix+".type", "")
		value, _ := m.cfg.GetString(condPrefix+".value", "")
		operator, _ := m.cfg.GetString(condPrefix+".operator", "equals")

		conditions = append(conditions, &Condition{
			Type:     condType,
			Value:    value,
			Operator: operator,
		})
	}

	return conditions
}

// loadRateLimit loads rate limit configuration
func (m *ActionManager) loadRateLimit(prefix string) *RateLimitConfig {
	requests, _ := m.cfg.GetInt(prefix+".requests", 10)
	window, _ := m.cfg.GetInt(prefix+".window", 60)
	scope, _ := m.cfg.GetString(prefix+".scope", "user")

	return &RateLimitConfig{
		Requests: requests,
		Window:   window,
		Scope:    scope,
	}
}

// StartScheduler starts the cron scheduler for scheduled actions
func (m *ActionManager) StartScheduler(ctx context.Context, session *discordgo.Session) {
	m.schedulerMu.Lock()
	defer m.schedulerMu.Unlock()

	for _, action := range m.actions {
		if action.Type == "scheduled" && action.Trigger.Schedule != "" {
			// Capture variables for closure
			act := action
			_, err := m.scheduler.AddFunc(act.Trigger.Schedule, func() {
				m.executeScheduledAction(ctx, session, act)
			})
			if err != nil {
				m.logger.Error("Failed to schedule action", "action", act.Name, "error", err)
			} else {
				m.logger.Info("Scheduled action", "action", act.Name, "schedule", act.Trigger.Schedule)
			}
		}
	}

	m.scheduler.Start()
}

// StopScheduler stops the cron scheduler
func (m *ActionManager) StopScheduler() {
	m.schedulerMu.Lock()
	defer m.schedulerMu.Unlock()

	if m.scheduler != nil {
		m.scheduler.Stop()
	}

	// Stop worker pool and wait for all tasks to complete
	if m.workerPool != nil {
		m.workerPool.StopAndWait()
		m.logger.Info("Worker pool stopped", "completedTasks", m.workerPool.SubmittedTasks())
	}
}

// executeScheduledAction executes a scheduled action
func (m *ActionManager) executeScheduledAction(ctx context.Context, s *discordgo.Session, action *Action) {
	// Send to configured channels
	for _, channelID := range action.Trigger.Channels {
		if err := m.sendResponse(ctx, s, channelID, "", action.Response); err != nil {
			m.logger.Error("Failed to execute scheduled action", "action", action.Name, "channel", channelID, "error", err)
		}
	}
}

// HandleMessage handles message events
func (m *ActionManager) HandleMessage(ctx context.Context, s *discordgo.Session, msg *discordgo.MessageCreate) {
	// Get command prefix
	prefix, _ := m.cfg.GetString("bot.prefix", "!")

	for _, action := range m.actions {
		// Skip non-message and non-command actions
		if action.Type != "message" && action.Type != "command" {
			continue
		}

		// Check if action matches
		matches := false
		if action.Type == "command" && action.Trigger.Command != "" {
			commandStr := prefix + action.Trigger.Command
			if len(msg.Content) >= len(commandStr) && msg.Content[:len(commandStr)] == commandStr {
				matches = true
			}
		} else if action.Type == "message" && action.Pattern != nil {
			if action.Pattern.MatchString(msg.Content) {
				matches = true
			}
		}

		if !matches {
			continue
		}

		// Check conditions
		if !m.checkConditions(ctx, s, msg, action) {
			continue
		}

		// Check rate limit
		if !m.checkRateLimit(msg.Author.ID, msg.ChannelID, msg.GuildID, action) {
			m.logger.Debug("Rate limit exceeded", "action", action.Name, "user", msg.Author.ID)
			continue
		}

		// Execute action in worker pool for concurrent processing
		act := action // Capture for closure
		m.workerPool.Submit(func() {
			if err := m.executeAction(ctx, s, msg, act); err != nil {
				m.logger.Error("Failed to execute action", "action", act.Name, "error", err)
			}
		})

		// Only execute first matching action
		break
	}
}

// HandleReaction handles reaction events
func (m *ActionManager) HandleReaction(ctx context.Context, s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	for _, action := range m.actions {
		if action.Type != "reaction" {
			continue
		}

		if action.Trigger.Emoji != r.Emoji.Name {
			continue
		}

		// Check rate limit
		if !m.checkRateLimit(r.UserID, r.ChannelID, r.GuildID, action) {
			continue
		}

		// Execute action in worker pool for concurrent processing
		act := action // Capture for closure
		m.workerPool.Submit(func() {
			if err := m.sendResponse(ctx, s, r.ChannelID, r.UserID, act.Response); err != nil {
				m.logger.Error("Failed to execute reaction action", "action", act.Name, "error", err)
			}
		})

		break
	}
}

// Continued in next file...
