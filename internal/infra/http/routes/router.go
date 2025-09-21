package routes

import (
	"zpmeow/docs"
	"zpmeow/internal/infra/http/handlers"
	"zpmeow/internal/infra/http/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// HandlerDependencies organized in the specified order:
// Health, Sessions, Messages, Privacy, Chat, Contacts, Groups, Communities, Newsletters, Webhooks
type HandlerDependencies struct {
	HealthHandler     *handlers.HealthHandler
	SessionHandler    *handlers.SessionHandler
	MessageHandler    *handlers.MessageHandler
	PrivacyHandler    *handlers.PrivacyHandler
	ChatHandler       *handlers.ChatHandler
	ContactHandler    *handlers.ContactHandler
	GroupHandler      *handlers.GroupHandler
	CommunityHandler  *handlers.CommunityHandler
	NewsletterHandler *handlers.NewsletterHandler
	WebhookHandler    *handlers.WebhookHandler
}

func SetupRoutes(
	app *fiber.App,
	handlers *HandlerDependencies,
	authMiddleware *middleware.AuthMiddleware,
) {

	// Recovery middleware is built into Fiber

	app.Use(func(c *fiber.Ctx) error {
		if c.Path() == "/swagger/doc.json" {
			host := c.Hostname()
			docs.SwaggerInfo.Host = host
		}
		return c.Next()
	})

	app.Get("/ping", handlers.HealthHandler.Ping)
	app.Get("/health", handlers.HealthHandler.Health)
	app.Get("/metrics", handlers.HealthHandler.Metrics)
	app.Post("/metrics/reset", handlers.HealthHandler.ResetMetrics)

	sessionGroup := app.Group("/sessions")
	sessionGroup.Use(authMiddleware.AuthenticateGlobal())
	sessionGroup.Post("/create", handlers.SessionHandler.CreateSession)
	sessionGroup.Get("/list", handlers.SessionHandler.GetSessions)
	sessionGroup.Get("/:sessionId/info", handlers.SessionHandler.GetSession)
	sessionGroup.Delete("/:sessionId/delete", handlers.SessionHandler.DeleteSession)
	sessionGroup.Post("/:sessionId/connect", handlers.SessionHandler.ConnectSession)
	sessionGroup.Post("/:sessionId/disconnect", handlers.SessionHandler.DisconnectSession)
	sessionGroup.Post("/:sessionId/pair", handlers.SessionHandler.PairPhone)
	sessionGroup.Get("/:sessionId/status", handlers.SessionHandler.GetSessionStatus)
	sessionGroup.Put("/:sessionId/webhook", handlers.SessionHandler.UpdateSessionWebhook)

	// Session API routes - TEMPORARILY COMMENTED OUT UNTIL HANDLERS ARE MIGRATED
	// TODO: Uncomment after migrating all handlers to Fiber
	/*
	sessionAPIGroup := app.Group("/session/:sessionId")
	sessionAPIGroup.Use(authMiddleware.AuthenticateSession())

	// 1. Messages
	sessionAPIGroup.Post("/message/send/text", handlers.MessageHandler.SendText)
	sessionAPIGroup.Post("/message/send/image", handlers.MessageHandler.SendImage)
	sessionAPIGroup.Post("/message/send/video", handlers.MessageHandler.SendVideo)
	sessionAPIGroup.Post("/message/send/audio", handlers.MessageHandler.SendAudio)
	sessionAPIGroup.Post("/message/send/document", handlers.MessageHandler.SendDocument)
	sessionAPIGroup.Post("/message/send/sticker", handlers.MessageHandler.SendSticker)
	sessionAPIGroup.Post("/message/send/contact", handlers.MessageHandler.SendContact)
	sessionAPIGroup.Post("/message/send/location", handlers.MessageHandler.SendLocation)
	sessionAPIGroup.Post("/message/send/media", handlers.MessageHandler.SendMedia)
	sessionAPIGroup.Post("/message/send/buttons", handlers.MessageHandler.SendButton)
	sessionAPIGroup.Post("/message/send/list", handlers.MessageHandler.SendList)
	sessionAPIGroup.Post("/message/send/poll", handlers.MessageHandler.SendPoll)

	sessionAPIGroup.Post("/message/markread", handlers.MessageHandler.MarkAsRead)
	sessionAPIGroup.Post("/message/react", handlers.MessageHandler.ReactToMessage)
	sessionAPIGroup.Post("/message/edit", handlers.MessageHandler.EditMessage)
	sessionAPIGroup.Post("/message/delete", handlers.MessageHandler.DeleteMessage)

	// 2. Privacy
	// TODO: Migrate PrivacyHandler to Fiber
	/*
	privacy := sessionAPIGroup.Group("/privacy")
	privacy.PUT("/set", handlers.PrivacyHandler.SetAllPrivacySettings)
	privacy.POST("/find", handlers.PrivacyHandler.FindPrivacySettings)
	privacy.GET("/blocklist", handlers.PrivacyHandler.GetBlocklist)
	privacy.PUT("/blocklist", handlers.PrivacyHandler.UpdateBlocklist)

	// 3. Chat
	// TODO: Migrate ChatHandler to Fiber
	chat := sessionAPIGroup.Group("/chat")
	chat.POST("/presence", handlers.ChatHandler.SetPresence)
	chat.GET("/history", handlers.ChatHandler.GetChatHistory)

	download := chat.Group("/download")
	download.POST("/image", handlers.ChatHandler.DownloadImage)
	download.POST("/video", handlers.ChatHandler.DownloadVideo)
	download.POST("/audio", handlers.ChatHandler.DownloadAudio)
	download.POST("/document", handlers.ChatHandler.DownloadDocument)

	chat.POST("/list", handlers.ChatHandler.ListChats)
	chat.POST("/info", handlers.ChatHandler.GetChatInfo)
	chat.POST("/pin", handlers.ChatHandler.PinChat)
	chat.POST("/mute", handlers.ChatHandler.MuteChat)
	chat.POST("/archive", handlers.ChatHandler.ArchiveChat)
	chat.POST("/disappearing-timer", handlers.ChatHandler.SetDisappearingTimer)

	// 4. Contacts
	// TODO: Migrate ContactHandler to Fiber
	contacts := sessionAPIGroup.Group("/contacts")
	contacts.POST("/check", handlers.ContactHandler.CheckUser)
	contacts.POST("/info", handlers.ContactHandler.GetUserInfo)
	contacts.POST("/avatar", handlers.ContactHandler.GetAvatar)
	contacts.GET("/list", handlers.ContactHandler.GetContacts)
	contacts.POST("/sync", handlers.ContactHandler.GetContacts)

	presence := sessionAPIGroup.Group("/presences")
	presence.PUT("/set", handlers.ContactHandler.SetPresence)
	presence.GET("/get", handlers.ContactHandler.GetUserInfo)
	presence.POST("/contact", handlers.ContactHandler.GetUserInfo)
	presence.POST("/subscribe", handlers.ContactHandler.CheckUser)
	presence.POST("/typing", handlers.ChatHandler.SetPresence)
	presence.POST("/recording", handlers.ChatHandler.SetPresence)

	// 5. Groups
	// TODO: Migrate GroupHandler to Fiber
	group := sessionAPIGroup.Group("/group")
	group.POST("/create", handlers.GroupHandler.CreateGroup)
	group.GET("/list", handlers.GroupHandler.ListGroups)
	group.POST("/info", handlers.GroupHandler.GetGroupInfo)
	group.POST("/join", handlers.GroupHandler.JoinGroup)
	group.POST("/join-with-invite", handlers.GroupHandler.JoinGroupWithInvite)
	group.POST("/leave", handlers.GroupHandler.LeaveGroup)
	group.POST("/invitelink", handlers.GroupHandler.GetInviteLink)
	group.POST("/inviteinfo", handlers.GroupHandler.GetInviteInfo)
	group.POST("/inviteinfo-specific", handlers.GroupHandler.GetGroupInfoFromInvite)

	participants := group.Group("/participants")
	participants.POST("/update", handlers.GroupHandler.UpdateParticipants)

	settings := group.Group("/settings")
	settings.POST("/name", handlers.GroupHandler.SetName)
	settings.POST("/topic", handlers.GroupHandler.SetTopic)
	settings.POST("/photo/set", handlers.GroupHandler.SetPhoto)
	settings.POST("/photo/remove", handlers.GroupHandler.RemovePhoto)
	settings.POST("/announce", handlers.GroupHandler.SetAnnounce)
	settings.POST("/locked", handlers.GroupHandler.SetLocked)
	settings.POST("/ephemeral", handlers.GroupHandler.SetEphemeral)
	settings.POST("/join-approval", handlers.GroupHandler.SetJoinApproval)
	settings.POST("/member-add-mode", handlers.GroupHandler.SetMemberAddMode)

	requests := group.Group("/requests")
	requests.POST("/list", handlers.GroupHandler.GetGroupRequestParticipants)
	requests.POST("/update", handlers.GroupHandler.UpdateGroupRequestParticipants)

	// 6. Communities
	// TODO: Migrate CommunityHandler to Fiber
	community := sessionAPIGroup.Group("/community")
	community.POST("/link", handlers.CommunityHandler.LinkGroup)
	community.POST("/unlink", handlers.CommunityHandler.UnlinkGroup)
	community.POST("/subgroups", handlers.CommunityHandler.GetSubGroups)
	community.POST("/participants", handlers.CommunityHandler.GetLinkedGroupsParticipants)

	// 7. Newsletters
	// TODO: Migrate NewsletterHandler to Fiber
	newsletter := sessionAPIGroup.Group("/newsletter")
	newsletter.POST("", handlers.NewsletterHandler.CreateNewsletter)
	newsletter.GET("/list", handlers.NewsletterHandler.ListNewsletters)
	newsletter.GET("/:newsletterId", handlers.NewsletterHandler.GetNewsletter)
	newsletter.POST("/:newsletterId/subscribe", handlers.NewsletterHandler.SubscribeToNewsletter)
	newsletter.POST("/:newsletterId/unsubscribe", handlers.NewsletterHandler.UnsubscribeFromNewsletter)
	newsletter.POST("/:newsletterId/send", handlers.NewsletterHandler.SendNewsletterMessage)

	newsletter.GET("/:newsletterId/messages", handlers.NewsletterHandler.GetNewsletterMessages)
	newsletter.GET("/:newsletterId/updates", handlers.NewsletterHandler.GetNewsletterMessageUpdates)
	newsletter.POST("/:newsletterId/mark-viewed", handlers.NewsletterHandler.MarkNewsletterViewed)

	newsletter.POST("/:newsletterId/reaction", handlers.NewsletterHandler.SendNewsletterReaction)
	newsletter.POST("/:newsletterId/mute", handlers.NewsletterHandler.ToggleNewsletterMute)
	newsletter.POST("/:newsletterId/live-updates", handlers.NewsletterHandler.SubscribeLiveUpdates)

	newsletter.POST("/upload", handlers.NewsletterHandler.UploadNewsletterMedia)
	newsletter.GET("/invite/:inviteKey", handlers.NewsletterHandler.GetNewsletterByInvite)

	// 8. Webhooks
	// TODO: Migrate WebhookHandler to Fiber
	webhook := sessionAPIGroup.Group("/webhook")
	webhook.POST("", handlers.WebhookHandler.SetWebhook)
	webhook.GET("", handlers.WebhookHandler.GetWebhook)

	webhooks := sessionAPIGroup.Group("/webhooks")
	webhooks.GET("/events", handlers.WebhookHandler.ListEvents)
	*/

	app.Get("/swagger/*", swagger.HandlerDefault)
}
