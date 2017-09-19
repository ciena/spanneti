package spanneti

import (
	"github.com/ciena/spanneti/spanneti/graph"
	"fmt"
	"github.com/pkg/errors"
	"reflect"
)

type Plugin struct {
	name          string
	startCallback func()
	eventCallback func(key string, value interface{})
	dataType      reflect.Type
}

func (spanneti *spanneti) getPluginData() map[string]reflect.Type {
	ret := make(map[string]reflect.Type)
	for name, plugin := range spanneti.plugins {
		ret[name] = plugin.dataType
	}
	return ret
}

func VerifyPlugin(plugin *Plugin) error {
	fmt.Printf("Verifying plugin '%s'...\n", plugin.name)

	//verify that the data type implements graph.PluginData
	if !plugin.dataType.Implements(reflect.TypeOf((*graph.PluginData)(nil)).Elem()) {
		return errors.New(plugin.dataType.Name() + " does not implement PluginData")
	}

	//verify keys that can be used for lookup
	paths, types, err := generatePaths("", plugin.dataType)
	if err != nil {
		return err
	}

	for i := 0; i < len(paths); i++ {
		fmt.Println("-", paths[i], types[i])
	}
	fmt.Println("Plugin OK")
	return nil
}

func generatePaths(path string, tipe reflect.Type) (paths []string, types []reflect.Type, _ error) {
	switch tipe.Kind() {
	case reflect.Struct:
		for i := 0; i < tipe.NumField(); i++ {
			if tipe.Field(i).Type.Kind() == reflect.Map && tipe.Field(i).Type.Key().Kind() == reflect.String {

				if tipe.Field(i).Type.Elem().Kind() == reflect.Struct {
					if addedPaths, addedTypes, err := generatePathsRecursive(path+tipe.Field(i).Name, tipe.Field(i).Type.Elem()); err != nil {
						return nil, nil, err
					} else {
						paths = append(paths, addedPaths...)
						types = append(types, addedTypes...)
					}
				}

				if tipe.Field(i).Tag.Get(TAG_NAME) != "" {
					paths = append(paths, tipe.Field(i).Tag.Get(TAG_NAME))
					types = append(types, tipe.Field(i).Type.Elem())
				}
			} else {
				if tipe.Field(i).Tag.Get(TAG_NAME) != "" {
					fmt.Printf("WARNING: Ignoring tag: %s; expected map[string]interface{} type\n", tipe.Field(i).Tag.Get(TAG_NAME))
				}
			}
		}
	}
	return
}

func generatePathsRecursive(path string, tipe reflect.Type) (paths []string, types []reflect.Type, _ error) {
	for i := 0; i < tipe.NumField(); i++ {

		if tipe.Field(i).Type.Kind() == reflect.Struct {
			if addedPaths, addedTypes, err := generatePathsRecursive(path+tipe.Field(i).Name, tipe.Field(i).Type); err != nil {
				return nil, nil, err
			} else {
				paths = append(paths, addedPaths...)
				types = append(types, addedTypes...)
			}
		}

		if tipe.Field(i).Tag.Get(TAG_NAME) != "" {
			paths = append(paths, tipe.Field(i).Tag.Get(TAG_NAME))
			types = append(types, tipe.Field(i).Type)
		}
	}
	return
}

const TAG_NAME = "spanneti"
