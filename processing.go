package main

import (
	"github.com/google/uuid"
)

type ServerStatusResponse struct {
	versionName     string
	versionProtocol int

	onlinePlayers *int
	maxPlayers    *int
	playerSample  *[]SamplePlayer
	isFakeSample  bool

	description *string
	favicon     *string

	enforcesSecureChat *bool
	previewsChat       *bool

	// non-vanilla fields
	// courtesy of https://github.com/mat-1/matscan/blob/e7f79b01ff36c2971b20948c575ad373651c8e77/src/processing/minecraft/mod.rs#L54
	// no chat reports and the like
	preventsChatReports *bool
	// forge
	forgedataFmlNetworkVersion *int
	// old forge servers
	modinfoType *string
	// neoforged
	isModded *bool
	// better compatibility checker
	modpackdataProjectId *int
	modpackdataName      *string
	modpackdataVersion   *string
}

type SamplePlayer struct {
	name string
	uuid uuid.UUID
}

// TODO implement
func ProcessJsonResponse(jsonStr string) (ServerStatusResponse, error) {
	serverStatusResponse := ServerStatusResponse{}
	return serverStatusResponse, nil
}
