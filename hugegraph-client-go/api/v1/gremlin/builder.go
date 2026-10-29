package gremlin

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Property 表示节点的单个属性键值对
type Property struct {
	Key   string
	Value any
}

// VertexInsertIfNotExistsBuilder 针对条件插入顶点 Gremlin 语句的构建器
type VertexInsertIfNotExistsBuilder struct {
	label      string
	id         any
	properties []Property
}

// NewVertexInsertIfNotExistsBuilder 创建一个 Builder 实例
func NewVertexInsertIfNotExistsBuilder(label string) *VertexInsertIfNotExistsBuilder {
	return &VertexInsertIfNotExistsBuilder{
		label:      label,
		properties: make([]Property, 0),
	}
}

// ID 设置顶点的 T.id
func (b *VertexInsertIfNotExistsBuilder) ID(id any) *VertexInsertIfNotExistsBuilder {
	b.id = id
	return b
}

// Property 添加单个属性（按添加顺序保留）
func (b *VertexInsertIfNotExistsBuilder) Property(key string, value any) *VertexInsertIfNotExistsBuilder {
	b.properties = append(b.properties, Property{Key: key, Value: value})
	return b
}

// Properties 批量追加属性
func (b *VertexInsertIfNotExistsBuilder) Properties(props map[string]any) *VertexInsertIfNotExistsBuilder {
	for k, v := range props {
		b.properties = append(b.properties, Property{Key: k, Value: v})
	}
	return b
}

// Build 生成完整的 Gremlin Groovy 脚本
func (b *VertexInsertIfNotExistsBuilder) Build() (string, error) {
	if strings.TrimSpace(b.label) == "" {
		return "", errors.New("label cannot be empty")
	}
	if b.id == nil || b.id == "" {
		return "", errors.New("id cannot be empty")
	}

	labelStr := formatValue(b.label)
	idStr := formatValue(b.id)

	// 1. 判断存在条件：g.V().hasLabel('...').has(T.id, '...').hasNext()
	condition := fmt.Sprintf("g.V().hasLabel(%s).has(T.id, %s).hasNext()", labelStr, idStr)

	// 2. 插入分支：g.addV('...').property(T.id, '...').property('k', 'v')...next(); 0
	var elseBranch strings.Builder
	elseBranch.WriteString(fmt.Sprintf("g.addV(%s).property(T.id, %s)", labelStr, idStr))

	for _, prop := range b.properties {
		elseBranch.WriteString(fmt.Sprintf(".property('%s', %s)", escapeGremlinString(prop.Key), formatValue(prop.Value)))
	}
	elseBranch.WriteString(".next(); 0")

	// 3. 组装整条语句
	script := fmt.Sprintf("if (%s) { 0 } else { %s }", condition, elseBranch.String())
	return script, nil
}

// formatValue 格式化 Go 变量为 Gremlin 字符串字面量
func formatValue(val any) string {
	switch v := val.(type) {
	case string:
		return fmt.Sprintf("'%s'", escapeGremlinString(v))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("'%s'", escapeGremlinString(fmt.Sprintf("%v", v)))
	}
}

// escapeGremlinString 转义字符串中的单引号与反斜杠，防止语法报错或注入
func escapeGremlinString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	return s
}
