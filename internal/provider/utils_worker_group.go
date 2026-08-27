package provider

// Kestra 2.0 replaced the namespace and tenant `workerGroup` reference — a single worker
// group addressed by key — with a `defaultWorkerSelector`: a tag set routed to matching
// Worker Queues, plus a matching strategy and a fallback. These helpers are shared by the
// namespace and tenant mappings.

func includedWorkerSelectorSchemaToApi(workerSelectorList []interface{}) map[string]interface{} {
	var workerSelectorData = make(map[string]interface{})

	if len(workerSelectorList) > 0 {
		workerSelectorMap := workerSelectorList[0].(map[string]interface{})

		tags := make([]string, 0)
		if rawTags, ok := workerSelectorMap["tags"].([]interface{}); ok {
			for _, tag := range rawTags {
				if s, ok := tag.(string); ok {
					tags = append(tags, s)
				}
			}
		}
		workerSelectorData["tags"] = tags

		if match, ok := workerSelectorMap["match"].(string); ok && match != "" {
			workerSelectorData["match"] = match
		}

		if fallback, ok := workerSelectorMap["fallback"].(string); ok && fallback != "" {
			workerSelectorData["fallback"] = fallback
		}
	}

	return workerSelectorData
}

func includedWorkerSelectorApiToList(workerSelector map[string]interface{}) []map[string]interface{} {
	var workerSelectorData = make(map[string]interface{})

	rawTags, ok := workerSelector["tags"].([]interface{})
	if !ok {
		return nil
	}
	tags := make([]string, 0, len(rawTags))
	for _, tag := range rawTags {
		if s, ok := tag.(string); ok {
			tags = append(tags, s)
		}
	}
	workerSelectorData["tags"] = tags

	if match, ok := workerSelector["match"].(string); ok {
		workerSelectorData["match"] = match
	}

	if fallback, ok := workerSelector["fallback"].(string); ok {
		workerSelectorData["fallback"] = fallback
	}

	var workerSelectorDataList []map[string]interface{}
	workerSelectorDataList = append(workerSelectorDataList, workerSelectorData)

	return workerSelectorDataList
}
