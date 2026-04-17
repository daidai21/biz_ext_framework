package biz_process

// Node is the common lightweight identity shared by process nodes.
type Node interface {
	NodeName() string
}
