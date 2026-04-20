package tool

type GetTimeReq struct {
	Location string `json:"location" jsonschema:"description=可选参数，用户询问时间的地点。如果不填则默认查询当前所在地的系统时间"`
}

type GetTimeResp struct {
	CurrentTime string `json:"current_time"`
}

type SearchKnowledgeReq struct {
	Query string `json:"query" jsonschema:"required" jsonschema_description:"用户问题的核心关键词或意图，用于检索私有知识库"`
}

type SearchKnowledgeResp struct {
	Context string `json:"context"`
}
