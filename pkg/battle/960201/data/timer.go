package data

type triggerEvent func(a string) (ok bool)
type TableTimer struct {
	TimerId     int          `json:"timer_id"`     //计数器的id
	TriggerTime int64        `json:"trigger_time"` //定时器出发的时间
	Event       triggerEvent //定时器要出发的事件，
	IsEffective bool         //是否有效，如果无效之后要从列表或者map中移除
	TimerName   string       //定时器唯一名称，
}
