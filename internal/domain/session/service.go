package session

type Service interface {
	CanConnect(session *Session) bool
	CanDisconnect(session *Session) bool
	CanDelete(session *Session) bool

	ValidateStatusTransition(current, newStatus Status) error

	ValidateSessionConfiguration(session *Session) error

	CanRegenerateApiKey(session *Session) bool

	CanSetProxy(session *Session) bool

	CanSubscribeToEvents(session *Session) bool

	ValidateDeviceConnection(session *Session, deviceJID string) error
}

type DomainService struct{}

func NewService() *DomainService {
	return &DomainService{}
}

func (s *DomainService) CanConnect(session *Session) bool {
	return session.IsDisconnected() || session.HasError() || session.IsConnecting()
}

func (s *DomainService) CanDisconnect(session *Session) bool {
	return session.IsConnected() || session.IsConnecting()
}

func (s *DomainService) CanDelete(session *Session) bool {
	return session.IsDisconnected()
}

func (s *DomainService) ValidateStatusTransition(current, newStatus Status) error {
	return ValidateSessionStatus(current, newStatus)
}

func (s *DomainService) ValidateSessionConfiguration(session *Session) error {
	if err := session.Validate(); err != nil {
		return err
	}

	if session.IsConnected() && !session.IsAuthenticated() {
		return ErrSessionNotConnected
	}

	return nil
}

func (s *DomainService) CanRegenerateApiKey(session *Session) bool {
	return !session.IsConnected()
}

func (s *DomainService) CanSetProxy(session *Session) bool {
	return !session.IsConnected()
}

func (s *DomainService) CanSubscribeToEvents(session *Session) bool {
	return true
}

func (s *DomainService) ValidateDeviceConnection(session *Session, deviceJID string) error {
	if session.IsConnected() && deviceJID == "" {
		return ErrSessionCannotBeConnectedWithoutDevice
	}

	if deviceJID == "" && session.IsConnected() {
		return ErrSessionCannotBeConnectedWithoutDevice
	}

	return nil
}

func ValidateSessionStatus(currentStatus, newStatus Status) error {
	validTransitions := map[Status][]Status{
		StatusDisconnected: {StatusConnecting},
		StatusConnecting:   {StatusConnected, StatusDisconnected, StatusError},
		StatusConnected:    {StatusDisconnected, StatusError},
		StatusError:        {StatusDisconnected, StatusConnecting},
	}

	allowedStatuses, exists := validTransitions[currentStatus]
	if !exists {
		return ErrInvalidSessionStatus
	}

	for _, allowed := range allowedStatuses {
		if newStatus == allowed {
			return nil
		}
	}

	return ErrInvalidSessionStatus
}
