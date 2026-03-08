package table

import (
	"github.com/busy-cloud/boat/javascript"
)

type Hook struct {
	BeforeInsert func(doc Document) error
	AfterInsert  func(id any, doc Document) error
	BeforeUpdate func(id any, update Document) error
	AfterUpdate  func(id any, update Document, base Document) error
	BeforeDelete func(id any) error
	AfterDelete  func(id any, doc Document) error

	Scripts *HookScripts `json:"scripts,omitempty"`
}

type HookScripts struct {
	BeforeInsert string `json:"before_insert,omitempty"`
	AfterInsert  string `json:"after_insert,omitempty"`
	BeforeUpdate string `json:"before_update,omitempty"`
	AfterUpdate  string `json:"after_update,omitempty"`
	BeforeDelete string `json:"before_delete,omitempty"`
	AfterDelete  string `json:"after_delete,omitempty"`
}

var globalValues = map[string]any{}

func SetHookValues(name string, value any) {
	globalValues[name] = value
}

func RemoveHookValues(name string) {
	delete(globalValues, name)
}

func (h *Hook) Compile() error {
	if h.Scripts == nil {
		return nil
	}

	if h.BeforeInsert == nil && h.Scripts.BeforeInsert != "" {
		program, err := javascript.Compile(h.Scripts.BeforeInsert)
		if err != nil {
			return err
		}
		h.BeforeInsert = func(doc Document) error {
			rt := javascript.Runtime()
			for k, v := range globalValues {
				_ = rt.Set(k, v)
			}
			_ = rt.Set(`document`, doc)
			_, err := rt.RunProgram(program)
			return err
		}
	}

	if h.AfterInsert == nil && h.Scripts.AfterInsert != "" {
		program, err := javascript.Compile(h.Scripts.AfterInsert)
		if err != nil {
			return err
		}
		h.AfterInsert = func(id any, doc Document) error {
			rt := javascript.Runtime()
			for k, v := range globalValues {
				_ = rt.Set(k, v)
			}
			_ = rt.Set(`id`, id)
			_ = rt.Set(`document`, doc)
			_, err := rt.RunProgram(program)
			return err
		}
	}

	if h.BeforeUpdate == nil && h.Scripts.BeforeUpdate != "" {
		program, err := javascript.Compile(h.Scripts.BeforeUpdate)
		if err != nil {
			return err
		}
		h.BeforeUpdate = func(id any, doc Document) error {
			rt := javascript.Runtime()
			for k, v := range globalValues {
				_ = rt.Set(k, v)
			}
			_ = rt.Set(`id`, id)
			_ = rt.Set(`document`, doc)
			_, err := rt.RunProgram(program)
			return err
		}
	}

	if h.AfterUpdate == nil && h.Scripts.AfterUpdate != "" {
		program, err := javascript.Compile(h.Scripts.AfterUpdate)
		if err != nil {
			return err
		}
		h.AfterUpdate = func(id any, update Document, base Document) error {
			rt := javascript.Runtime()
			for k, v := range globalValues {
				_ = rt.Set(k, v)
			}
			_ = rt.Set(`id`, id)
			_ = rt.Set(`update`, update)
			_ = rt.Set(`document`, base)
			_, err := rt.RunProgram(program)
			return err
		}
	}

	if h.BeforeDelete == nil && h.Scripts.BeforeDelete != "" {
		program, err := javascript.Compile(h.Scripts.BeforeDelete)
		if err != nil {
			return err
		}
		h.BeforeDelete = func(id any) error {
			rt := javascript.Runtime()
			for k, v := range globalValues {
				_ = rt.Set(k, v)
			}
			_ = rt.Set(`id`, id)
			_, err := rt.RunProgram(program)
			return err
		}
	}

	if h.AfterDelete == nil && h.Scripts.AfterDelete != "" {
		program, err := javascript.Compile(h.Scripts.AfterDelete)
		if err != nil {
			return err
		}
		h.AfterDelete = func(id any, doc Document) error {
			rt := javascript.Runtime()
			for k, v := range globalValues {
				_ = rt.Set(k, v)
			}
			_ = rt.Set(`id`, id)
			_ = rt.Set(`document`, doc)
			_, err := rt.RunProgram(program)
			return err
		}
	}

	return nil
}
