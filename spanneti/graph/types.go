package graph

import (
	"fmt"
	"reflect"
)

type ContainerID string

type ContainerNetwork struct {
	PluginData  map[string]PluginData
	ContainerId ContainerID
}

type PluginData interface {
	SetContainerID(containerId ContainerID) PluginData
}

func (containerNet *ContainerNetwork) KeyValueMap() map[string]map[string][]interface{} {
	ret := make(map[string]map[string][]interface{})
	for pluginName, object := range containerNet.PluginData {
		if _, have := ret[pluginName]; !have {
			ret[pluginName] = make(map[string][]interface{})
		}

		keys, values := generatePaths(object)
		for i := 0; i < len(keys); i++ {
			ret[pluginName][keys[i]] = append(ret[pluginName][keys[i]], values[i])
		}
	}
	return ret
}

const TAG_NAME = "spanneti"

func generatePaths(object interface{}) (keys []string, values []interface{}) {
	tipe := reflect.TypeOf(object)
	value := reflect.ValueOf(object)
	switch value.Kind() {
	case reflect.Struct:
		for i := 0; i < tipe.NumField(); i++ {

			if value.Field(i).Kind() == reflect.Map && value.Field(i).Type().Key().Kind() == reflect.String {
				if value.Field(i).Len() > 0 {
					for _, mapKey := range value.Field(i).MapKeys() {
						mapValue := value.Field(i).MapIndex(mapKey)

						if mapValue.Kind() == reflect.Struct {
							addedKeys, addedValues := generatePathsRecursive(mapValue)
							keys = append(keys, addedKeys...)
							values = append(values, addedValues...)
						}

						if tipe.Field(i).Tag.Get(TAG_NAME) != "" {
							keys = append(keys, tipe.Field(i).Tag.Get(TAG_NAME))
							values = append(values, mapValue.Interface())
						}
					}

				}
			} else {
				if tipe.Field(i).Tag.Get(TAG_NAME) != "" {
					fmt.Printf("WARNING: Ignoring tag: '%s'; expected map[string]interface{} type\n", tipe.Field(i).Tag.Get(TAG_NAME))
				}
			}
		}
	}
	return
}

func generatePathsRecursive(value reflect.Value) (keys []string, values []interface{}) {
	for i := 0; i < value.Type().NumField(); i++ {

		if value.Field(i).Type().Kind() == reflect.Struct {
			addedPaths, addedTypes := generatePathsRecursive(value.Field(i))
			keys = append(keys, addedPaths...)
			values = append(values, addedTypes...)
		}

		if tagName := value.Type().Field(i).Tag.Get(TAG_NAME); tagName != "" {
			keys = append(keys, tagName)
			values = append(values, value.Field(i).Interface())
		}
	}
	return
}
