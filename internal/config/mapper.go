package config

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// ConfigMapper 自动配置映射器
type ConfigMapper struct {
	endpointToConfig map[string]reflect.Type
}

var globalMapper = &ConfigMapper{
	endpointToConfig: map[string]reflect.Type{
		"youdao":     reflect.TypeOf(YoudaoConfig{}),
		"etymonline": reflect.TypeOf(EtymonlineConfig{}),
		"ecdict":     reflect.TypeOf(EcdictConfig{}),
		"mwebster":   reflect.TypeOf(MWebsterConfig{}),
		"llm":        reflect.TypeOf(LLMConfig{}),
	},
}

// MapToEndpointConfig 将parameters自动映射到指定endpoint的配置结构体
func (cm *ConfigMapper) MapToEndpointConfig(endpoint string, parameters map[string]interface{}) (DictEndpointConfig, error) {
	configType, exists := cm.endpointToConfig[endpoint]
	if !exists {
		return nil, fmt.Errorf("unknown endpoint: %s", endpoint)
	}

	// 创建配置实例
	configValue := reflect.New(configType).Elem()
	config := configValue.Addr().Interface().(DictEndpointConfig)

	return cm.mapToConfigValue(config, configValue, parameters)
}

// MapToConfig 将parameters自动映射到指定配置结构体
func (cm *ConfigMapper) MapToConfig(config DictEndpointConfig, parameters map[string]interface{}) (DictEndpointConfig, error) {
	configValue := reflect.ValueOf(config)
	if configValue.Kind() != reflect.Ptr || configValue.Elem().Kind() != reflect.Struct {
		return nil, errors.New("config must be a pointer to struct")
	}
	return cm.mapToConfigValue(config, configValue.Elem(), parameters)
}

func (cm *ConfigMapper) mapToConfigValue(config DictEndpointConfig, configValue reflect.Value, parameters map[string]interface{}) (DictEndpointConfig, error) {
	// 获取默认值
	defaults := config.GetDefaults()
	var defaultsMap map[string]interface{}
	if defaults != nil {
		defaultsMap = defaults.(map[string]interface{})
	} else {
		defaultsMap = make(map[string]interface{})
	}

	mergedParams := cm.mergeDefaults(parameters, defaultsMap)

	// 自动映射字段
	if err := cm.mapFields(configValue, mergedParams); err != nil {
		return nil, err
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// mapFields 使用反射自动映射字段
func (cm *ConfigMapper) mapFields(configValue reflect.Value, parameters map[string]interface{}) error {
	configType := configValue.Type()

	for i := 0; i < configValue.NumField(); i++ {
		field := configValue.Field(i)
		fieldStruct := configType.Field(i)

		// 获取YAML标签
		yamlTag := fieldStruct.Tag.Get("yaml")
		if yamlTag == "" || yamlTag == "-" {
			continue
		}

		// 查找参数值
		paramValue, exists := parameters[yamlTag]
		if !exists {
			continue
		}

		// 类型转换和赋值
		if err := cm.setFieldValue(field, paramValue); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to set field %s (yaml: %s)", fieldStruct.Name, yamlTag))
		}
	}

	return nil
}

// setFieldValue 处理类型转换和字段赋值
func (cm *ConfigMapper) setFieldValue(field reflect.Value, value interface{}) error {
	if !field.CanSet() {
		return errors.New("field is not settable")
	}

	fieldKind := field.Kind()
	valueKind := reflect.TypeOf(value).Kind()

	switch fieldKind {
	case reflect.String:
		if valueKind == reflect.String {
			field.SetString(value.(string))
			return nil
		}
		return fmt.Errorf("expected string, got %T", value)

	case reflect.Int:
		switch valueKind {
		case reflect.Int:
			field.SetInt(int64(value.(int)))
			return nil
		case reflect.Float64:
			field.SetInt(int64(value.(float64)))
			return nil
		case reflect.String:
			if intValue, err := strconv.Atoi(value.(string)); err == nil {
				field.SetInt(int64(intValue))
				return nil
			}
			return fmt.Errorf("cannot convert string %q to int", value)
		}
		return fmt.Errorf("expected int, got %T", value)

	case reflect.Float64:
		switch valueKind {
		case reflect.Float64:
			field.SetFloat(value.(float64))
			return nil
		case reflect.String:
			if floatValue, err := strconv.ParseFloat(value.(string), 64); err == nil {
				field.SetFloat(floatValue)
				return nil
			}
			return fmt.Errorf("cannot convert string %q to float64", value)
		}
		return fmt.Errorf("expected float64, got %T", value)

	default:
		// 特殊处理time.Duration
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			switch valueKind {
			case reflect.Int:
				field.Set(reflect.ValueOf(time.Duration(value.(int))))
				return nil
			case reflect.String:
				if duration, err := time.ParseDuration(value.(string)); err == nil {
					field.Set(reflect.ValueOf(duration))
					return nil
				}
				return fmt.Errorf("cannot parse duration from string %q", value)
			}
			return fmt.Errorf("expected duration, got %T", value)
		}
		return fmt.Errorf("unsupported field type: %s", fieldKind)
	}
}

// mergeDefaults 合并默认值和用户参数
func (cm *ConfigMapper) mergeDefaults(parameters, defaults map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// 先添加默认值
	for k, v := range defaults {
		merged[k] = v
	}

	// 覆盖用户参数
	for k, v := range parameters {
		merged[k] = v
	}

	return merged
}
