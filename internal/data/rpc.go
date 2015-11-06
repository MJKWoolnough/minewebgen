package data

type DefaultMap struct {
	Mode               int
	Name               string
	GameMode           int32
	Seed               int64
	Structures, Cheats bool
}

type SuperFlatMap struct {
	DefaultMap
	GeneratorSettings string
}

type CustomMap struct {
	DefaultMap
	GeneratorSettings struct {
		SeaLevel                uint8   `json:"seaLevel"`
		Caves                   bool    `json:"useCaves"`
		Strongholds             bool    `json:"useStrongholds"`
		Villages                bool    `json:"useVillages"`
		Mineshafts              bool    `json:"useMineShafts"`
		Temples                 bool    `json:"useTemples"`
		OceanMonuments          bool    `json:"useMonuments"` // Needs checking
		Ravines                 bool    `json:"useRavines"`
		Dungeons                bool    `json:"useDungeons"`
		DungeonChance           uint8   `json:"dungeonChance"`
		WaterLake               bool    `json:"useWaterLake"`
		WaterLakeChance         uint8   `json:"waterLakeChance"`
		LaveLake                bool    `json:"useLavaLake"`
		LavaLakeChance          uint8   `json:"lavaLakeChance"`
		LavaOceans              bool    `json:"useLavaOceans"`
		Biome                   int16   `json:"fixedBiome"`
		BiomeSize               uint8   `json:"biomeSize"`
		RiverSize               uint8   `json:"riverSize"`
		DirtSize                uint8   `json:"dirtSize"`
		DirtTries               uint8   `json:"dirtCount"`
		DirtMinHeight           uint8   `json:"dirtMinHeight"`
		DirtMaxHeight           uint8   `json:"dirtMaxHeight"`
		GravelSize              uint8   `json:"gravelSize"`
		GravelTries             uint8   `json:"gravelCount"`
		GravelMinHeight         uint8   `json:"gravelMinHeight"`
		GravelMaxHeight         uint8   `json:"gravelMaxHeight"`
		GraniteSize             uint8   `json:"graniteSize"`
		GraniteTries            uint8   `json:"graniteCount"`
		GraniteMinHeight        uint8   `json:"graniteMinHeight"`
		GraniteMaxHeight        uint8   `json:"graniteMaxHeight"`
		DiortiteSize            uint8   `json:"dioriteSize"`
		DiortiteTries           uint8   `json:"dioriteCount"`
		DiortiteMinHeight       uint8   `json:"dioriteMinHeight"`
		DiortiteMaxHeight       uint8   `json:"dioriteMaxHeight"`
		AndesiteSize            uint8   `json:"andesiteSize"`
		AndesiteTries           uint8   `json:"andesiteCount"`
		AndesiteMinHeight       uint8   `json:"andesiteMinHeight"`
		AndesiteMaxHeight       uint8   `json:"andesiteMaxHeight"`
		CoalSize                uint8   `json:"coalSize"`
		CoalTries               uint8   `json:"coalCount"`
		CoalMinHeight           uint8   `json:"coalMinHeight"`
		CoalMaxHeight           uint8   `json:"coalMaxHeight"`
		IronSize                uint8   `json:"ironSize"`
		IronTries               uint8   `json:"ironCount"`
		IronMinHeight           uint8   `json:"ironMinHeight"`
		IronMaxHeight           uint8   `json:"ironMaxHeight"`
		GoldSize                uint8   `json:"goldSize"`
		GoldTries               uint8   `json:"goldCount"`
		GoldMinHeight           uint8   `json:"goldMinHeight"`
		GoldMaxHeight           uint8   `json:"goldMaxHeight"`
		RedstoneSize            uint8   `json:"redstoneSize"`
		RedstoneTries           uint8   `json:"redstoneCount"`
		RedstoneMinHeight       uint8   `json:"redstoneMinHeight"`
		RedstoneMaxHeight       uint8   `json:"redstoneMaxHeight"`
		DiamondSize             uint8   `json:"diamondSize"`
		DiamondTries            uint8   `json:"diamondCount"`
		DiamondMinHeight        uint8   `json:"diamondMinHeight"`
		DiamondMaxHeight        uint8   `json:"diamondMaxHeight"`
		LapisSize               uint8   `json:"lapisSize"`
		LapisTries              uint8   `json:"lapisCount"`
		LapisCenterHeight       uint8   `json:"lapisCenterHeight"`
		LapisSpread             uint8   `json:"lapisSpread"`
		MainNoiseScaleX         float32 `json:"mainNoiseScaleX"`
		MainNoiseScaleY         float32 `json:"mainNoiseScaleY"`
		MainNoiseScaleZ         float32 `json:"mainNoiseScaleZ"`
		DepthNoiseScaleX        float32 `json:"depthNoiseScaleX"`
		DepthNoiseScaleY        float32 `json:"depthNoiseScaleY"`
		DepthNoiseScaleZ        float32 `json:"depthNoiseScaleZ"`
		DepthNoiseScaleExponent float32 `json:"depthNoiseScaleExponent"`
		BaseSize                float32 `json:"baseSize"`
		CoordinateScale         float32 `json:"coordinateScale"`
		HeightScale             float32 `json:"heightScale"`
		HeightStretch           float32 `json:"stretchY"`
		UpperLimitScale         float32 `json:"upperLimitScale"`
		LowerLimitScale         float32 `json:"lowerLimitScale"`
		BiomeDepthWeight        float32 `json:"biomeDepthWeight"`
		BiomeScaleOffset        float32 `json:"biomeDepthOffset"`
		BiomeScaleWeight        float32 `json:"biomeScaleWeight"`
		BiomeDepthOffset        float32 `json:"biomeScaleOffset"`
	}
}
