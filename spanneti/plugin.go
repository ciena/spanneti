package spanneti

import (
	"bitbucket.ciena.com/BP_ONOS/spanneti/spanneti/graph"
	"fmt"
	"reflect"
)

type Plugin interface {
	Name() string
	Start(spanneti Spanneti)
	Event(key string, value interface{})
	DataType() reflect.Type
}

func (spanneti *spanneti) getPluginData() map[string]reflect.Type {
	ret := make(map[string]reflect.Type)
	for name, plugin := range spanneti.plugins {
		ret[name] = plugin.DataType()
	}
	return ret
}

func VerifyPlugins(plugins []Plugin) {
	fmt.Println("Verifying plugins...")
	for _, plugin := range plugins {
		fmt.Println(plugin.Name())

		//verify that the data type implements graph.PluginData
		if !plugin.DataType().Implements(reflect.TypeOf((*graph.PluginData)(nil)).Elem()) {
			panic(plugin.DataType().Name() + " does not implement PluginData")
		}

		//verify keys that can be used for lookup
		paths, types, err := generatePaths("", plugin.DataType())
		if err != nil {
			panic(err)
		}

		for i := 0; i < len(paths); i++ {
			fmt.Println("-", paths[i], types[i])
		}
	}
	fmt.Println("Plugins OK\n")
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
