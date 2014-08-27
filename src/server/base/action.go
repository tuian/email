package base

import (
	"database/sql"
)

type Action interface {
	Exec(email *EMail, args ...interface{}) error
}

type LabelAction struct{}
type ForwardAction struct{}
type ReplyAction struct{}
type MoveMessaeAction struct{}
type CopyMessageAction struct{}
type ChangeStatusAction struct{}

// 照着Mac Outlook里面的Rules支持的动作写的
// 摘抄了部分容易实现的功能
const (
	kActionMoveMessage     = "MoveMessage"
	kActionCopyMessage     = "CopyMessage"
	kActionDeleteMessage   = "DeleteMessage"
	kActionLabel           = "Label"
	kActionChangeStatus    = "ChangeStatus"
	kActionToDo            = "ToDo"
	kActionReply           = "Reply"
	kActionForward         = "Forward"
	kActionSaveAttachments = "SaveAttachments"
	kActionDoNotNotify     = "DoNotNotify"
)

// 给邮件打Tag
func (this LabelAction) Exec(email *EMail, args ...interface{}) error {
	value := args[0]
	switch value.(type) {
	case string:
		label := value.(string)
		db := args[1].(*sql.DB)
		return email.AddLabel(label, db)
	}
	return nil
}

func (this ForwardAction) Exec(email *EMail, args ...interface{}) error {
	return nil
}

func (this ReplyAction) Exec(email *EMail, args ...interface{}) error {
	return nil
}

func (this MoveMessaeAction) Exec(email *EMail, args ...interface{}) error {
	return nil
}

func (this CopyMessageAction) Exec(email *EMail, args ...interface{}) error {
	return nil
}

func (this ChangeStatusAction) Exec(email *EMail, args ...interface{}) error {
	return nil
}

// 过滤器要执行的动作
func NewAction(t string) Action {
	switch t {
	case kActionLabel:
		return LabelAction{}
	}
	return nil
}