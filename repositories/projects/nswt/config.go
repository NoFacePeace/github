package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func loadRushFlowOptions(path string) (rushFlowOptions, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return rushFlowOptions{}, err
	}

	options := rushFlowOptions{
		CollectType: 10,
	}
	if err := json.Unmarshal(data, &options); err != nil {
		return rushFlowOptions{}, fmt.Errorf("decode config: %w", err)
	}

	return options, nil
}
