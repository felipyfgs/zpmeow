package routes

import (
	"zpmeow/docs"
	"zpmeow/internal/infra/http/handlers"
	"zpmeow/internal/infra/http/middleware"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
)

type HandlerDependencies struct {
	HealthHandler     *handlers.HealthHandler
	SessionHandler    *handlers.SessionHandler
	ContactHandler    *handlers.ContactHandler
	ChatHandler       *handlers.ChatHandler
	MessageHandler    *handlers.MessageHandler
	GroupHandler      *handlers.GroupHandler
	CommunityHandler  *handlers.CommunityHandler
	NewsletterHandler *handlers.NewsletterHandler
	WebhookHandler    *handlers.WebhookHandler
	PrivacyHandler    *handlers.PrivacyHandler
}

func SetupRoutes(
	router *gin.Engine,
	handlers *HandlerDependencies,
	authMiddleware *middleware.AuthMiddleware,
) {

	router.Use(middleware.Logger())
	router.Use(gin.Recovery())

	router.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/swagger/doc.json" {
			host := c.Request.Host
			docs.SwaggerInfo.Host = host
		}
		c.Next()
	})

	router.GET("/ping", handlers.HealthHandler.Ping)
	router.GET("/health", handlers.HealthHandler.Health)
	router.GET("/metrics", handlers.HealthHandler.Metrics)
	router.POST("/metrics/reset", handlers.HealthHandler.ResetMetrics)

	sessionGroup := router.Group("/sessions")
	sessionGroup.Use(authMiddleware.AuthenticateGlobal())
	{
		sessionGroup.POST("/create", handlers.SessionHandler.CreateSession)
		sessionGroup.GET("/list", handlers.SessionHandler.GetSessions)
		sessionGroup.GET("/:sessionId/info", handlers.SessionHandler.GetSession)
		sessionGroup.DELETE("/:sessionId/delete", handlers.SessionHandler.DeleteSession)
		sessionGroup.POST("/:sessionId/connect", handlers.SessionHandler.ConnectSession)
		sessionGroup.POST("/:sessionId/disconnect", handlers.SessionHandler.DisconnectSession)
		sessionGroup.POST("/:sessionId/pair", handlers.SessionHandler.PairPhone)
		sessionGroup.GET("/:sessionId/status", handlers.SessionHandler.GetSessionStatus)
		sessionGroup.PUT("/:sessionId/webhook", handlers.SessionHandler.UpdateSessionWebhook)
	}

	sessionAPIGroup := router.Group("/session/:sessionId")
	sessionAPIGroup.Use(authMiddleware.AuthenticateSession())
	{
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

		privacy := sessionAPIGroup.Group("/privacy")
		privacy.PUT("/set", handlers.PrivacyHandler.SetAllPrivacySettings)
		privacy.POST("/find", handlers.PrivacyHandler.FindPrivacySettings)
		privacy.GET("/blocklist", handlers.PrivacyHandler.GetBlocklist)
		privacy.PUT("/blocklist", handlers.PrivacyHandler.UpdateBlocklist)

		message := sessionAPIGroup.Group("/message")
		send := message.Group("/send")
		send.POST("/text", handlers.MessageHandler.SendText)
		send.POST("/image", handlers.MessageHandler.SendImage)
		send.POST("/video", handlers.MessageHandler.SendVideo)
		send.POST("/audio", handlers.MessageHandler.SendAudio)
		send.POST("/document", handlers.MessageHandler.SendDocument)
		send.POST("/sticker", handlers.MessageHandler.SendSticker)
		send.POST("/contact", handlers.MessageHandler.SendContact)
		send.POST("/location", handlers.MessageHandler.SendLocation)
		send.POST("/media", handlers.MessageHandler.SendMedia)
		send.POST("/buttons", handlers.MessageHandler.SendButton)
		send.POST("/list", handlers.MessageHandler.SendList)
		send.POST("/poll", handlers.MessageHandler.SendPoll)

		message.POST("/markread", handlers.MessageHandler.MarkAsRead)
		message.POST("/react", handlers.MessageHandler.ReactToMessage)
		message.POST("/edit", handlers.MessageHandler.EditMessage)
		message.POST("/delete", handlers.MessageHandler.DeleteMessage)

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

		community := sessionAPIGroup.Group("/community")
		community.POST("/link", handlers.CommunityHandler.LinkGroup)
		community.POST("/unlink", handlers.CommunityHandler.UnlinkGroup)
		community.POST("/subgroups", handlers.CommunityHandler.GetSubGroups)
		community.POST("/participants", handlers.CommunityHandler.GetLinkedGroupsParticipants)

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

		webhook := sessionAPIGroup.Group("/webhook")
		webhook.POST("", handlers.WebhookHandler.SetWebhook)
		webhook.GET("", handlers.WebhookHandler.GetWebhook)

		webhooks := sessionAPIGroup.Group("/webhooks")
		webhooks.GET("/events", handlers.WebhookHandler.ListEvents)
	}

	router.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler, ginswagger.URL("/swagger/doc.json")))
}
