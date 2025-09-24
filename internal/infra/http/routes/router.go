package routes

import (
	"zpmeow/docs"
	"zpmeow/internal/infra/http/handlers"
	"zpmeow/internal/infra/http/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

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
	ChatwootHandler   *handlers.ChatwootHandler
}

func SetupRoutes(
	app *fiber.App,
	handlers *HandlerDependencies,
	authMiddleware *middleware.AuthMiddleware,
) {

	app.Use(func(c *fiber.Ctx) error {
		if c.Path() == "/swagger/doc.json" {
			host := c.Hostname()
			docs.SwaggerInfo.Host = host
		}
		return c.Next()
	})

	app.Get("/health", handlers.HealthHandler.Health)

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

	sessionAPIGroup := app.Group("/session/:sessionId")
	sessionAPIGroup.Use(authMiddleware.AuthenticateSession())

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

	privacy := sessionAPIGroup.Group("/privacy")
	privacy.Put("/set", handlers.PrivacyHandler.SetAllPrivacySettings)
	privacy.Post("/find", handlers.PrivacyHandler.FindPrivacySettings)
	privacy.Get("/blocklist", handlers.PrivacyHandler.GetBlocklist)
	privacy.Put("/blocklist", handlers.PrivacyHandler.UpdateBlocklist)

	contacts := sessionAPIGroup.Group("/contacts")
	contacts.Post("/check", handlers.ContactHandler.CheckUser)
	contacts.Post("/info", handlers.ContactHandler.GetUserInfo)
	contacts.Post("/avatar", handlers.ContactHandler.GetAvatar)
	contacts.Get("/list", handlers.ContactHandler.GetContacts)
	contacts.Post("/sync", handlers.ContactHandler.GetContacts)

	presence := sessionAPIGroup.Group("/presences")
	presence.Put("/set", handlers.ContactHandler.SetPresence)
	presence.Get("/get", handlers.ContactHandler.GetUserInfo)
	presence.Post("/contact", handlers.ContactHandler.GetUserInfo)
	presence.Post("/subscribe", handlers.ContactHandler.CheckUser)

	chat := sessionAPIGroup.Group("/chat")
	chat.Post("/presence", handlers.ChatHandler.SetPresence)
	chat.Get("/history", handlers.ChatHandler.GetChatHistory)

	download := chat.Group("/download")
	download.Post("/image", handlers.ChatHandler.DownloadImage)
	download.Post("/video", handlers.ChatHandler.DownloadVideo)
	download.Post("/audio", handlers.ChatHandler.DownloadAudio)
	download.Post("/document", handlers.ChatHandler.DownloadDocument)

	chat.Post("/list", handlers.ChatHandler.ListChats)
	chat.Post("/info", handlers.ChatHandler.GetChatInfo)
	chat.Post("/pin", handlers.ChatHandler.PinChat)
	chat.Post("/mute", handlers.ChatHandler.MuteChat)
	chat.Post("/archive", handlers.ChatHandler.ArchiveChat)
	chat.Post("/disappearing-timer", handlers.ChatHandler.SetDisappearingTimer)

	chatPresence := sessionAPIGroup.Group("/presences")
	chatPresence.Post("/typing", handlers.ChatHandler.SetPresence)
	chatPresence.Post("/recording", handlers.ChatHandler.SetPresence)

	group := sessionAPIGroup.Group("/group")
	group.Post("/create", handlers.GroupHandler.CreateGroup)
	group.Get("/list", handlers.GroupHandler.ListGroups)
	group.Post("/info", handlers.GroupHandler.GetGroupInfo)
	group.Post("/join", handlers.GroupHandler.JoinGroup)
	group.Post("/join-with-invite", handlers.GroupHandler.JoinGroupWithInvite)
	group.Post("/leave", handlers.GroupHandler.LeaveGroup)
	group.Post("/invitelink", handlers.GroupHandler.GetInviteLink)
	group.Post("/inviteinfo", handlers.GroupHandler.GetInviteInfo)
	group.Post("/inviteinfo-specific", handlers.GroupHandler.GetGroupInfoFromInvite)

	participants := group.Group("/participants")
	participants.Post("/update", handlers.GroupHandler.UpdateParticipants)

	settings := group.Group("/settings")
	settings.Post("/name", handlers.GroupHandler.SetName)
	settings.Post("/topic", handlers.GroupHandler.SetTopic)
	settings.Post("/photo/set", handlers.GroupHandler.SetPhoto)
	settings.Post("/photo/remove", handlers.GroupHandler.RemovePhoto)
	settings.Post("/announce", handlers.GroupHandler.SetAnnounce)
	settings.Post("/locked", handlers.GroupHandler.SetLocked)
	settings.Post("/ephemeral", handlers.GroupHandler.SetEphemeral)
	settings.Post("/join-approval", handlers.GroupHandler.SetJoinApproval)
	settings.Post("/member-add-mode", handlers.GroupHandler.SetMemberAddMode)

	requests := group.Group("/requests")
	requests.Post("/list", handlers.GroupHandler.GetGroupRequestParticipants)
	requests.Post("/update", handlers.GroupHandler.UpdateGroupRequestParticipants)

	community := sessionAPIGroup.Group("/community")
	community.Post("/link", handlers.CommunityHandler.LinkGroup)
	community.Post("/unlink", handlers.CommunityHandler.UnlinkGroup)
	community.Post("/subgroups", handlers.CommunityHandler.GetSubGroups)
	community.Post("/participants", handlers.CommunityHandler.GetLinkedGroupsParticipants)

	newsletter := sessionAPIGroup.Group("/newsletter")
	newsletter.Post("", handlers.NewsletterHandler.CreateNewsletter)
	newsletter.Get("/list", handlers.NewsletterHandler.ListNewsletters)
	newsletter.Get("/:newsletterId", handlers.NewsletterHandler.GetNewsletter)
	newsletter.Post("/:newsletterId/subscribe", handlers.NewsletterHandler.SubscribeToNewsletter)
	newsletter.Post("/:newsletterId/unsubscribe", handlers.NewsletterHandler.UnsubscribeFromNewsletter)
	newsletter.Post("/:newsletterId/send", handlers.NewsletterHandler.SendNewsletterMessage)

	newsletter.Get("/:newsletterId/messages", handlers.NewsletterHandler.GetNewsletterMessages)
	newsletter.Get("/:newsletterId/updates", handlers.NewsletterHandler.GetNewsletterMessageUpdates)
	newsletter.Post("/:newsletterId/mark-viewed", handlers.NewsletterHandler.MarkNewsletterViewed)

	newsletter.Post("/:newsletterId/reaction", handlers.NewsletterHandler.SendNewsletterReaction)
	newsletter.Post("/:newsletterId/mute", handlers.NewsletterHandler.ToggleNewsletterMute)
	newsletter.Post("/:newsletterId/live-updates", handlers.NewsletterHandler.SubscribeLiveUpdates)

	newsletter.Post("/upload", handlers.NewsletterHandler.UploadNewsletterMedia)
	newsletter.Get("/invite/:inviteKey", handlers.NewsletterHandler.GetNewsletterByInvite)

	webhook := sessionAPIGroup.Group("/webhook")
	webhook.Post("", handlers.WebhookHandler.SetWebhook)
	webhook.Get("", handlers.WebhookHandler.GetWebhook)

	webhooks := sessionAPIGroup.Group("/webhooks")
	webhooks.Get("/events", handlers.WebhookHandler.ListEvents)

	// Chatwoot integration routes (simplified)
	chatwoot := sessionAPIGroup.Group("/chatwoot")
	chatwoot.Post("/set", handlers.ChatwootHandler.SetChatwootConfig)
	chatwoot.Get("/find", handlers.ChatwootHandler.GetChatwootConfig)

	// Chatwoot webhook route (internal, not in swagger - used by Chatwoot to send data)
	app.Post("/chatwoot/webhook/:sessionId", handlers.ChatwootHandler.ReceiveChatwootWebhook)

	app.Get("/swagger/swagger-config.json", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"tagsSorter":       nil,
			"operationsSorter": nil,
		})
	})

	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL:                      "/swagger/doc.json",
		ConfigURL:                "/swagger/swagger-config.json",
		DeepLinking:              true,
		DocExpansion:             "list",
		DefaultModelsExpandDepth: 1,
		DefaultModelExpandDepth:  1,
	}))
}
