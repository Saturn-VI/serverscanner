package main

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type ServerStatusDTO struct {
	Version            VersionInfo  `json:"version"`
	Players            *PlayersInfo `json:"players,omitempty"`
	Description        string       `json:"description"`
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

	IsFakeSample bool
	IsOnlineMode bool
}

type VersionInfo struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}

type PlayersInfo struct {
	Max    int            `json:"max"`
	Online int            `json:"online"`
	Sample []SamplePlayer `json:"sample,omitempty"`
}

type SamplePlayer struct {
	Name string    `json:"name"`
	ID   uuid.UUID `json:"id"`
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

// TODO implement
func ProcessJsonResponse(jsonStr string) (ServerStatus, error) {
	ssDTO := &ServerStatusDTO{}
	json.Unmarshal([]byte(jsonStr), ssDTO)

	fmt.Println(ssDTO)

	serverStatus := ServerStatus{
		ServerStatusDTO: *ssDTO,
		// TODO determine these values
		IsFakeSample: false,
		IsOnlineMode: true,
	}
	return serverStatus, nil
}
