package smart

/**
智能表单

类型
  text:
  password:
  number:
  slider:
  radio:
  rate:
  select:
  tags:
  color:
  checkbox:
  switch:
  textarea:
  date:
  time:
  datetime:
  file:
  image:
  images:
  object:
  list:
  table:
*/

type Field struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Type        string `json:"type,omitempty"` //type object array
	Default     any    `json:"default,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Tips        string `json:"tips,omitempty"`

	Clear    bool `json:"clear,omitempty"`
	Disabled bool `json:"disabled,omitempty"`
	Hidden   bool `json:"hidden,omitempty"`

	Array    bool    `json:"array,omitempty"`
	Children []Field `json:"children,omitempty"` //子级？

	Required bool    `json:"required,omitempty"`
	Min      float64 `json:"min,omitempty"`
	Max      float64 `json:"max,omitempty"`
	Step     float64 `json:"step,omitempty"`

	Multiple bool `json:"multiple,omitempty"`

	Auto    []AutoOption   `json:"auto,omitempty"`
	Options []SelectOption `json:"options,omitempty"`
	Tree    []TreeOption   `json:"tree,omitempty"`

	Change string `json:"change,omitempty"`

	Time       bool   `json:"time,omitempty"`
	TimeFormat string `json:"time_format,omitempty"`

	Pattern string `json:"pattern,omitempty"`

	Upload string `json:"upload,omitempty"` //上传路径

	//TODO 这里要及时补充，与前端保持一致
}

type AutoOption struct {
	Label string `json:"label"`
	Value any    `json:"value"`
}

type SelectOption struct {
	Label    string `json:"label"`
	Value    any    `json:"value"`
	Title    string `json:"title,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
	Hide     bool   `json:"hide,omitempty"`
	Key      any    `json:"key,omitempty"`
}

type TreeOption struct {
	Title      string        `json:"title,omitempty"`
	Key        string        `json:"key,omitempty"`
	IsLeaf     bool          `json:"isLeaf,omitempty"`
	Selectable bool          `json:"selectable,omitempty"`
	Disabled   bool          `json:"disabled,omitempty"`
	Expanded   bool          `json:"expanded,omitempty"`
	Children   []*TreeOption `json:"children,omitempty"`
}
