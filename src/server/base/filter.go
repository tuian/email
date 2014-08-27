package base

import (
	"bytes"
	// "fmt"
	"io/ioutil"
	"log"
	"path"

	"../net/mail"
)

const (
	kOpIs             = "Is"
	kOpContains       = "Contains"
	kOpExists         = "Exists"
	KOpOnlyMe         = "Only Me"
	kOpMe             = "Me"
	kOpNotMe          = "Not Me"
	kOpCcMe           = "Cc Me"
	kOpToOrCcMe       = "To or Cc Me"
	kOpHasAttachments = "HasAttachments"
	kOpRange          = "Range"
	kMatchAll         = "All"
	kMatchAny         = "Any"
)

const (
	kHFrom          = "From"
	kHTo            = "To"
	kHCc            = "Cc"
	kHSubject       = "Subject"
	kHBody          = "Body"
	kHSubjectOrBody = "SubjectOrBody"
	kHDate          = "Date"
	kHSize          = "Size"
	kHAttachments   = "Attachments"
)

type Filter struct {
	Name      string
	Disable   bool
	Stop      bool
	Condition struct {
		Match string
		Rules [][]string
	}
	Action map[string]string
}

func CheckRule(email *EMail, msg *mail.Message, rule []string) bool {
	key := rule[0]

	operator := NewOperator(rule[1])
	if operator == nil {
		log.Fatal("Can't support such operator = [" + rule[1] + "]")
		return false
	}

	switch key {
	case kHFrom:
		from, err := mail.ParseAddress(email.From)
		if err != nil {
			return false
		}

		return operator.Exec(from.Address, rule[2])
	case kHTo:
		to, err := mail.ParseAddressList(email.To)
		if err != nil {
			return false
		}

		return operator.Exec(to, rule[2])
	case kHCc:
		cc, err := mail.ParseAddressList(email.Cc)
		if err != nil {
			return false
		}

		return operator.Exec(cc, rule[2])
	case kHSubject:
		return operator.Exec(email.Subject, rule[2])
	case kHBody:
		return operator.Exec(email.Message, rule[2])
	case kHSubjectOrBody:
		a := operator.Exec(email.Subject, rule[2])
		b := operator.Exec(email.Message, rule[2])
		return a || b
	case kHDate:
		return operator.Exec(email.Date, rule[2])
	// case kHSize:
	// 	break
	case kHAttachments:
		// TODO(user) 这个判断准确么
		a := msg.Header.Get("X-Has-Attach") == "yes"
		b := msg.Header.Get("X-MS-Has-Attach") == "yes"
		return a || b
	default:
		return operator.Exec(msg.Header.Get(key), rule[2])
	}

	return false
}

// 判断邮件是否符合规则
func (filter *Filter) Match(email *EMail) bool {
	raw, err := ioutil.ReadFile(path.Join("raw", email.Uidl+".txt"))
	if err != nil {
		log.Fatal(err)
		return false
	}

	msg, err := mail.ReadMessage(bytes.NewBuffer(raw))
	if err != nil {
		log.Fatal(err)
		return false
	}

	match := filter.Condition.Match
	if match == kMatchAll {
		// 全部条件都要满足
		for _, rule := range filter.Condition.Rules {
			if len(rule) <= 1 {
				// 至少需要2项
				continue
			}

			if !CheckRule(email, msg, rule[:]) {
				return false
			}
		}

		return true
	} else if match == kMatchAny {
		// 满足任意一个条件即可
		for _, rule := range filter.Condition.Rules {
			if len(rule) <= 1 {
				// 至少需要2项
				continue
			}

			if CheckRule(email, msg, rule[:]) {
				return true
			}
		}

		return false
	} else {
		// TODO(user) 暂不支持
		return false
	}
}

// 如果符合规则的话，也就是通过Match判断的话，执行规则定义的动作
func (filter *Filter) TakeAction(email *EMail) error {
	return nil
}

// 针对一封邮件，运行一边所有的Filter
func RunFilter(email *EMail, filters []Filter) error {
	for _, filter := range filters {
		if filter.Disable {
			continue
		}

		if filter.Match(email) {
			err := filter.TakeAction(email)
			if err != nil {
				return err
			}

			if filter.Stop {
				return nil
			}
		}
	}

	return nil
}