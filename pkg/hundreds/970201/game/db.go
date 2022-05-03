package game

//
////初始化数据表
//func InitTable() {
//	orm.ConnectUserDb(config.GameFrameConfig.DbUser, config.GameFrameConfig.DbPwd, config.GameFrameConfig.DbIp, config.GameFrameConfig.DbName)
//	if orm.GetDb() == nil {
//		panic("orm db is nil ")
//	}
//	orm.GetDb().AutoMigrate(&TableRed{})
//	orm.GetDb().AutoMigrate(&TableRedRob{})
//	log.Traceln("数据库连接成功")
//}
//
//func GetTableRedById(redId int64) (tableRed *TableRed) {
//	tableRed = new(TableRed)
//	orm.GetDb().Where("id=?", redId).First(tableRed)
//	return
//}
//
//func SaveTableRed(instance *TableRed) (err error) {
//	err = orm.GetDb().Save(instance).Error
//	if err != nil {
//		log.Traceln("SaveTableRed err : ", err)
//		return
//	}
//	return
//}
//
//type TableRedList struct {
//	List        []*TableRed
//	Pager       *page.Pager
//	TotalAmount int64 `json:"total_amount"` //累加之后的总额
//}
//
////获取发过的红包列表-分页
//func GetTableRedList(condition interface{}, pageIndex, pageSize int) (res *TableRedList) {
//	list := make([]*TableRed, 0)
//	// total count
//	var count int
//	//whereSql := CombineWhereOr(condition)
//	orm.GetDb().Model(TableRed{}).Where(condition).Count(&count)
//	pager := page.NewPager(pageIndex, pageSize, count)
//
//	// database paging
//	err := orm.GetDb().Debug().Model(TableRed{}).Where(condition).
//		Order("id desc").Offset(pager.Begin).Limit(pageSize).Find(&list).Error
//	if err != nil {
//		return
//	}
//	ac := new(AmountCount)
//	orm.GetDb().Model(TableRed{}).Where(condition).Select("SUM(amount) AS `total_amount`").Scan(&ac)
//	res = &TableRedList{
//		List: list, Pager: pager,TotalAmount:ac.TotalAmount,
//	}
//	return
//}
//
//func SaveTableRedRob(instance *TableRedRob) (err error) {
//	err = orm.GetDb().Save(instance).Error
//	return
//}
//
//type TableRedRobList struct {
//	List        []*TableRedRob
//	Pager       *page.Pager
//	TotalAmount int64 `json:"total_amount"` //累加之后的总额
//}
//
////获取抢过的红包列表-分页
//func GetTableRedRobList(condition interface{}, pageIndex, pageSize int) (res *TableRedRobList) {
//	list := make([]*TableRedRob, 0)
//	var count int
//	orm.GetDb().Model(TableRedRob{}).Where(condition).Count(&count)
//	pager := page.NewPager(pageIndex, pageSize, count)
//
//	// database paging
//	orm.GetDb().Debug().Model(TableRedRob{}).Where(condition).
//		Order("id desc").Offset(pager.Begin).Limit(pageSize).Find(&list)
//	ac := new(AmountCount)
//	orm.GetDb().Model(TableRedRob{}).Where(condition).Select("SUM(robbed_amount) AS `total_amount`").Scan(&ac)
//	res = &TableRedRobList{
//		List: list, Pager: pager,TotalAmount:ac.TotalAmount,
//	}
//	return
//}
//
//type AmountCount struct {
//	TotalAmount int64 `json:"total_amount"`
//	TotalCount  int64 `json:"total_count"`
//}
//
////获取某个红包中雷赔付的全部金额
//func GetTableMineTotalAmount(redId int64) (ac *AmountCount) {
//	ac = new(AmountCount)
//	orm.GetDb().Where("red_id=?",redId).Select("SUM(mine_amount) AS `total_amount`").
//	Scan(&ac)
//	return
//}
//
////获取用户中雷总额
//func GetUserMineSum(uid int64) *AmountCount {
//	ac := new(AmountCount)
//	orm.GetDb().Model(TableRedRob{}).Where(TableRedRob{Uid:uid,IsMine:true}).
//		Select("SUM(mine_amount) AS `total_amount`").Scan(&ac)
//	return ac
//}
