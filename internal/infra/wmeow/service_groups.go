package wmeow

import (
	"context"
	"fmt"
	"strings"

	"zpmeow/internal/application/ports"

	"go.mau.fi/whatsmeow"
	waTypes "go.mau.fi/whatsmeow/types"
)

// GroupManager methods - gestão completa de grupos

func (m *MeowService) CreateGroup(ctx context.Context, sessionID, name string, participants []string) (*ports.GroupInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	var participantJIDs []waTypes.JID
	for _, participant := range participants {
		jid, err := parsePhoneToJID(participant)
		if err != nil {
			m.logger.Warnf("Invalid participant phone number %s: %v", participant, err)
			continue
		}
		participantJIDs = append(participantJIDs, jid)
	}

	if len(participantJIDs) == 0 {
		return nil, fmt.Errorf("no valid participants provided")
	}

	groupInfo, err := client.GetClient().CreateGroup(whatsmeow.ReqCreateGroup{
		Name:         name,
		Participants: participantJIDs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	result := &ports.GroupInfo{
		JID:         groupInfo.JID.String(),
		Name:        name,
		Participants: participants,
		CreatedAt:   groupInfo.CreateTime,
	}

	m.logger.Debugf("Group '%s' created successfully: %s for session %s",
		name, groupInfo.JID.String(), sessionID)

	return result, nil
}

func (m *MeowService) ListGroups(ctx context.Context, sessionID string) ([]ports.GroupInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	groups, err := client.GetClient().GetJoinedGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get joined groups: %w", err)
	}

	var result []ports.GroupInfo
	for _, group := range groups {
		if group == nil {
			continue
		}

		groupInfo := ports.GroupInfo{
			JID:       group.JID.String(),
			Name:      group.Name,
			CreatedAt: group.GroupCreated,
		}

		// Get participants
		for _, participant := range group.Participants {
			if participant.JID.Server == waTypes.DefaultUserServer {
				phone := participant.JID.User
				if !strings.HasPrefix(phone, "+") {
					phone = "+" + phone
				}
				groupInfo.Participants = append(groupInfo.Participants, phone)
			}
		}

		result = append(result, groupInfo)
	}

	m.logger.Debugf("Retrieved %d groups for session %s", len(result), sessionID)
	return result, nil
}

func (m *MeowService) GetGroupInfo(ctx context.Context, sessionID, groupJID string) (*ports.GroupInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	groupInfo, err := client.GetClient().GetGroupInfo(jid)
	if err != nil {
		return nil, fmt.Errorf("failed to get group info: %w", err)
	}

	result := &ports.GroupInfo{
		JID:         groupInfo.JID.String(),
		Name:        groupInfo.Name,
		Description: groupInfo.Topic,
		CreatedAt:   groupInfo.GroupCreated,
	}

	// Get participants
	for _, participant := range groupInfo.Participants {
		if participant.JID.Server == waTypes.DefaultUserServer {
			phone := participant.JID.User
			if !strings.HasPrefix(phone, "+") {
				phone = "+" + phone
			}
			result.Participants = append(result.Participants, phone)
		}
	}

	return result, nil
}

func (m *MeowService) JoinGroup(ctx context.Context, sessionID, inviteLink string) (*ports.GroupInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	groupInfo, err := client.GetClient().JoinGroupWithLink(inviteLink)
	if err != nil {
		return nil, fmt.Errorf("failed to join group: %w", err)
	}

	result := &ports.GroupInfo{
		JID:       groupInfo.JID.String(),
		Name:      groupInfo.Name,
		CreatedAt: groupInfo.GroupCreated,
	}

	return result, nil
}

func (m *MeowService) JoinGroupWithInvite(ctx context.Context, sessionID, groupJID, inviter, code string, expiration int64) (*ports.GroupInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	inviterJID, err := waTypes.ParseJID(inviter)
	if err != nil {
		return nil, fmt.Errorf("invalid inviter JID %s: %w", inviter, err)
	}

	groupInfo, err := client.GetClient().JoinGroupWithInvite(jid, inviterJID, code, expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to join group with invite: %w", err)
	}

	result := &ports.GroupInfo{
		JID:       groupInfo.JID.String(),
		Name:      groupInfo.Name,
		CreatedAt: groupInfo.GroupCreated,
	}

	return result, nil
}

func (m *MeowService) LeaveGroup(ctx context.Context, sessionID, groupJID string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	_, err = client.GetClient().LeaveGroup(jid)
	if err != nil {
		return fmt.Errorf("failed to leave group: %w", err)
	}

	m.logger.Debugf("Left group %s for session %s", groupJID, sessionID)
	return nil
}

func (m *MeowService) GetInviteLink(ctx context.Context, sessionID, groupJID string, reset bool) (string, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return "", fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return "", fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return "", fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	link, err := client.GetClient().GetGroupInviteLink(jid, reset)
	if err != nil {
		return "", fmt.Errorf("failed to get group invite link: %w", err)
	}

	return link, nil
}

func (m *MeowService) GetGroupInviteLink(ctx context.Context, sessionID, groupJID string) (string, error) {
	return m.GetInviteLink(ctx, sessionID, groupJID, false)
}

func (m *MeowService) AddParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	var participantJIDs []waTypes.JID
	for _, participant := range participants {
		participantJID, err := parsePhoneToJID(participant)
		if err != nil {
			m.logger.Warnf("Invalid participant phone number %s: %v", participant, err)
			continue
		}
		participantJIDs = append(participantJIDs, participantJID)
	}

	if len(participantJIDs) == 0 {
		return fmt.Errorf("no valid participants provided")
	}

	_, err = client.GetClient().UpdateGroupParticipants(jid, participantJIDs, whatsmeow.ParticipantChangeAdd)
	if err != nil {
		return fmt.Errorf("failed to add participants: %w", err)
	}

	m.logger.Debugf("Added %d participants to group %s for session %s", len(participantJIDs), groupJID, sessionID)
	return nil
}

func (m *MeowService) RemoveParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	var participantJIDs []waTypes.JID
	for _, participant := range participants {
		participantJID, err := parsePhoneToJID(participant)
		if err != nil {
			m.logger.Warnf("Invalid participant phone number %s: %v", participant, err)
			continue
		}
		participantJIDs = append(participantJIDs, participantJID)
	}

	if len(participantJIDs) == 0 {
		return fmt.Errorf("no valid participants provided")
	}

	_, err = client.GetClient().UpdateGroupParticipants(jid, participantJIDs, whatsmeow.ParticipantChangeRemove)
	if err != nil {
		return fmt.Errorf("failed to remove participants: %w", err)
	}

	m.logger.Debugf("Removed %d participants from group %s for session %s", len(participantJIDs), groupJID, sessionID)
	return nil
}

func (m *MeowService) PromoteParticipant(ctx context.Context, sessionID, groupJID string, participants []string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	var participantJIDs []waTypes.JID
	for _, participant := range participants {
		participantJID, err := parsePhoneToJID(participant)
		if err != nil {
			m.logger.Warnf("Invalid participant phone number %s: %v", participant, err)
			continue
		}
		participantJIDs = append(participantJIDs, participantJID)
	}

	if len(participantJIDs) == 0 {
		return fmt.Errorf("no valid participants provided")
	}

	_, err = client.GetClient().UpdateGroupParticipants(jid, participantJIDs, whatsmeow.ParticipantChangePromote)
	if err != nil {
		return fmt.Errorf("failed to promote participants: %w", err)
	}

	m.logger.Debugf("Promoted %d participants in group %s for session %s", len(participantJIDs), groupJID, sessionID)
	return nil
}

func (m *MeowService) DemoteParticipant(ctx context.Context, sessionID, groupJID string, participants []string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	var participantJIDs []waTypes.JID
	for _, participant := range participants {
		participantJID, err := parsePhoneToJID(participant)
		if err != nil {
			m.logger.Warnf("Invalid participant phone number %s: %v", participant, err)
			continue
		}
		participantJIDs = append(participantJIDs, participantJID)
	}

	if len(participantJIDs) == 0 {
		return fmt.Errorf("no valid participants provided")
	}

	_, err = client.GetClient().UpdateGroupParticipants(jid, participantJIDs, whatsmeow.ParticipantChangeDemote)
	if err != nil {
		return fmt.Errorf("failed to demote participants: %w", err)
	}

	m.logger.Debugf("Demoted %d participants in group %s for session %s", len(participantJIDs), groupJID, sessionID)
	return nil
}

func (m *MeowService) UpdateGroupName(ctx context.Context, sessionID, groupJID, name string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	err = client.GetClient().SetGroupName(jid, name)
	if err != nil {
		return fmt.Errorf("failed to update group name: %w", err)
	}

	m.logger.Debugf("Updated group name to '%s' for group %s in session %s", name, groupJID, sessionID)
	return nil
}

func (m *MeowService) UpdateGroupDescription(ctx context.Context, sessionID, groupJID, description string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	err = client.GetClient().SetGroupTopic(jid, description)
	if err != nil {
		return fmt.Errorf("failed to update group description: %w", err)
	}

	m.logger.Debugf("Updated group description for group %s in session %s", groupJID, sessionID)
	return nil
}

func (m *MeowService) SetGroupPhoto(ctx context.Context, sessionID, groupJID string, photoData []byte) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	_, err = client.GetClient().SetGroupPhoto(jid, photoData)
	if err != nil {
		return fmt.Errorf("failed to set group photo: %w", err)
	}

	m.logger.Debugf("Set group photo for group %s in session %s", groupJID, sessionID)
	return nil
}

func (m *MeowService) GetGroupRequestParticipants(ctx context.Context, sessionID, groupJID string) ([]string, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	participants, err := client.GetClient().GetGroupRequestParticipants(jid)
	if err != nil {
		return nil, fmt.Errorf("failed to get group request participants: %w", err)
	}

	var result []string
	for _, participant := range participants {
		if participant.JID.Server == waTypes.DefaultUserServer {
			phone := participant.JID.User
			if !strings.HasPrefix(phone, "+") {
				phone = "+" + phone
			}
			result = append(result, phone)
		}
	}

	m.logger.Debugf("Retrieved %d pending participants for group %s in session %s", len(result), groupJID, sessionID)
	return result, nil
}

func (m *MeowService) GetSubGroups(ctx context.Context, sessionID, communityJID string) ([]string, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	communityJIDParsed, err := waTypes.ParseJID(communityJID)
	if err != nil {
		return nil, fmt.Errorf("invalid community JID %s: %w", communityJID, err)
	}

	subGroups, err := client.GetClient().GetSubGroups(communityJIDParsed)
	if err != nil {
		return nil, fmt.Errorf("failed to get subgroups: %w", err)
	}

	var groups []string
	for _, group := range subGroups {
		if group != nil {
			groups = append(groups, group.JID.String())
		}
	}

	m.logger.Debugf("Retrieved %d subgroups for community %s in session %s", len(groups), communityJID, sessionID)
	return groups, nil
}

func (m *MeowService) GetLinkedGroupsParticipants(ctx context.Context, sessionID, communityJID string) ([]string, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	communityJIDParsed, err := waTypes.ParseJID(communityJID)
	if err != nil {
		return nil, fmt.Errorf("invalid community JID %s: %w", communityJID, err)
	}

	participants, err := client.GetClient().GetLinkedGroupsParticipants(communityJIDParsed)
	if err != nil {
		return nil, fmt.Errorf("failed to get linked groups participants: %w", err)
	}

	var result []string
	for _, participant := range participants {
		if participant.JID.Server == waTypes.DefaultUserServer {
			phone := participant.JID.User
			if !strings.HasPrefix(phone, "+") {
				phone = "+" + phone
			}
			result = append(result, phone)
		}
	}

	m.logger.Debugf("Retrieved %d linked groups participants for community %s in session %s", len(result), communityJID, sessionID)
	return result, nil
}

// Additional methods required by GroupManager interface

func (m *MeowService) GetInviteInfo(ctx context.Context, sessionID, inviteLink string) (*ports.GroupInfo, error) {
	// TODO: Implement GetInviteInfo
	return nil, fmt.Errorf("GetInviteInfo not implemented yet")
}

func (m *MeowService) GetGroupInfoFromInvite(ctx context.Context, sessionID, groupJID, inviter, code string, expiration int64) (*ports.GroupInfo, error) {
	// TODO: Implement GetGroupInfoFromInvite
	return nil, fmt.Errorf("GetGroupInfoFromInvite not implemented yet")
}

func (m *MeowService) UpdateParticipants(ctx context.Context, sessionID, groupJID, action string, participants []string) error {
	switch action {
	case "add":
		return m.AddParticipants(ctx, sessionID, groupJID, participants)
	case "remove":
		return m.RemoveParticipants(ctx, sessionID, groupJID, participants)
	case "promote":
		return m.PromoteParticipant(ctx, sessionID, groupJID, participants)
	case "demote":
		return m.DemoteParticipant(ctx, sessionID, groupJID, participants)
	default:
		return fmt.Errorf("invalid action: %s", action)
	}
}

func (m *MeowService) SetGroupName(ctx context.Context, sessionID, groupJID, name string) error {
	return m.UpdateGroupName(ctx, sessionID, groupJID, name)
}

func (m *MeowService) SetGroupTopic(ctx context.Context, sessionID, groupJID, topic string) error {
	return m.UpdateGroupDescription(ctx, sessionID, groupJID, topic)
}

func (m *MeowService) RemoveGroupPhoto(ctx context.Context, sessionID, groupJID string) error {
	// TODO: Implement RemoveGroupPhoto
	return fmt.Errorf("RemoveGroupPhoto not implemented yet")
}

func (m *MeowService) SetGroupAnnounce(ctx context.Context, sessionID, groupJID string, announceOnly bool) error {
	// TODO: Implement SetGroupAnnounce
	return fmt.Errorf("SetGroupAnnounce not implemented yet")
}

func (m *MeowService) SetGroupLocked(ctx context.Context, sessionID, groupJID string, locked bool) error {
	// TODO: Implement SetGroupLocked
	return fmt.Errorf("SetGroupLocked not implemented yet")
}

func (m *MeowService) SetGroupEphemeral(ctx context.Context, sessionID, groupJID string, ephemeral bool, duration int) error {
	// TODO: Implement SetGroupEphemeral
	return fmt.Errorf("SetGroupEphemeral not implemented yet")
}

func (m *MeowService) SetGroupJoinApprovalMode(ctx context.Context, sessionID, groupJID string, requireApproval bool) error {
	// TODO: Implement SetGroupJoinApprovalMode
	return fmt.Errorf("SetGroupJoinApprovalMode not implemented yet")
}

func (m *MeowService) SetGroupJoinApproval(ctx context.Context, sessionID, groupJID string, requireApproval bool) error {
	// TODO: Implement SetGroupJoinApproval
	return fmt.Errorf("SetGroupJoinApproval not implemented yet")
}

func (m *MeowService) SetGroupMemberAddMode(ctx context.Context, sessionID, groupJID string, mode string) error {
	// TODO: Implement SetGroupMemberAddMode
	return fmt.Errorf("SetGroupMemberAddMode not implemented yet")
}
