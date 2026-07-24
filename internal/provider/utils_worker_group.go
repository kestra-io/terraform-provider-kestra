package provider

func includedWorkerGroupSchemaToApi(workerGroupList []interface{}) map[string]interface{} {
	var workerGroupData = make(map[string]interface{})

	if len(workerGroupList) > 0 {
		workerGroupMap := workerGroupList[0].(map[string]interface{})
		workerGroupData["key"] = workerGroupMap["key"].(string)

		if workerGroupMap["fallback"] != "" {
			workerGroupData["fallback"] = workerGroupMap["fallback"].(string)
		}
	}

	return workerGroupData
}

func includedWorkerGroupApiToList(workerGroup map[string]interface{}) []map[string]interface{} {
	var workerGroupData = make(map[string]interface{})

	if key, ok := workerGroup["key"].(string); ok {
		workerGroupData["key"] = key
	} else {
		return nil
	}

	if workerGroup["fallback"] != nil {
		workerGroupData["fallback"] = workerGroup["fallback"].(string)
	}

	var workerGroupDataList []map[string]interface{}
	workerGroupDataList = append(workerGroupDataList, workerGroupData)

	return workerGroupDataList
}
