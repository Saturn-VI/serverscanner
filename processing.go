package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
)

const FAKE_SAMPLE_DESCRIPTION = "To protect the privacy of this server and its\nusers, you must log in once to see ping data."
const ANONYMOUS_PLAYER_NAME = "Anonymous Player"

type Description string

func (d *Description) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*d = Description(s)
		return nil
	}

	var obj struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		*d = Description(string(obj.Text))
		return nil
	}

	return fmt.Errorf("invalid description format")
}

type ServerStatusDTO struct {
	Version            VersionInfo  `json:"version"`
	Players            *PlayersInfo `json:"players,omitempty"`
	Description        Description  `json:"description"`
	Favicon            *string      `json:"favicon,omitempty"`
	EnforcesSecureChat *bool        `json:"enforcesSecureChat,omitempty"`
	PreviewsChat       *bool        `json:"previewsChat,omitempty"`

	// non-vanilla fields
	// courtesy of https://github.com/mat-1/matscan/blob/e7f79b01ff36c2971b20948c575ad373651c8e77/src/processing/minecraft/mod.rs#L54
	// in order:
	// no chat reports and the like (PreventsChatReports)
	// forge (ForgeDataInfo)
	// old forge servers (ModinfoType)
	// neoforged (IsModded)
	// better compatibility checker (ModpackData)
	PreventsChatReports *bool            `json:"preventsChatReports,omitempty"`
	ForgeDataInfo       *ForgeDataInfo   `json:"forgeData,omitempty"`
	ModinfoType         *ModInfo         `json:"modinfo,omitempty"`
	IsModded            *bool            `json:"isModded,omitempty"`
	ModpackData         *ModpackDataInfo `json:"modpackData,omitempty"`
}

type ServerStatus struct {
	ServerStatusDTO
	Address net.Addr  `json:"addr,omitempty"`
	Time    time.Time `json:"time,omitempty"`

	// IsFakeSample should take precedence over IsOnlineMode
	IsFakeSample bool  `json:"isFakeSample"`
	IsOnlineMode *bool `json:"isOnlineMode,omitempty"`
}

type VersionInfo struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}

type PlayersInfo struct {
	Max    int             `json:"max"`
	Online int             `json:"online"`
	Sample *[]SamplePlayer `json:"sample,omitempty"`
}

type SamplePlayer struct {
	Name *string    `json:"name,omitempty"`
	ID   *uuid.UUID `json:"id,omitempty"`
}

type ForgeDataInfo struct {
	int `json:"fmlNetworkVersion"`
}

type ModInfo struct {
	Type string `json:"type"`
}

type ModpackDataInfo struct {
	ProjectID int    `json:"projectID"`
	Name      string `json:"name"`
	Version   string `json:"version"`
}

func ProcessJsonResponse(jsonStr string, addr net.Addr) (*ServerStatus, error) {
	ssDTO := &ServerStatusDTO{}
	err := json.Unmarshal([]byte(jsonStr), ssDTO)
	if err != nil {
		return nil, err
	}

	isFake := false
	var isOnline *bool = nil

	if ssDTO.Description == FAKE_SAMPLE_DESCRIPTION {
		isFake = true
	}

	if ssDTO.Players == nil || ssDTO.Players.Sample == nil {
		// no sample means we can't determine if it's online or offline mode
		return &ServerStatus{
			ServerStatusDTO: *ssDTO,
			Address:         addr,
			Time:            time.Now(),

			IsFakeSample: true,
			IsOnlineMode: nil,
		}, nil
	}

	seenUUIDs := make(map[uuid.UUID]bool)
	for _, player := range *ssDTO.Players.Sample {
		// must have name and uuid
		if player.Name == nil || player.ID == nil {
			isFake = true
			break
		}
		// no duplicate uuids
		if _, exists := seenUUIDs[*player.ID]; exists {
			isFake = true
			break
		}
		seenUUIDs[*player.ID] = true

		id, err := uuid.Parse(player.ID.String())
		if err != nil {
			if *player.Name == ANONYMOUS_PLAYER_NAME {
				// allow anonymous player with invalid UUID
				continue
			}
			isFake = true
			break
		}
		switch id.Version() {
		case 4:
			// online mode
			isOnline = newTrue()
		case 3:
			// offline mode
			if isOnline == nil {
				isOnline = newFalse()
			}
		default:
			isFake = true
		}
	}

	serverStatus := ServerStatus{
		ServerStatusDTO: *ssDTO,
		Address:         addr,
		Time:            time.Now(),

		IsFakeSample: isFake,
		IsOnlineMode: isOnline,
	}
	return &serverStatus, nil
}
