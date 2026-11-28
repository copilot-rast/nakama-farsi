package configtransfer

func t2(fa, en string) map[string]string { return map[string]string{"fa": fa, "en": en} }

var (
	catKey = []string{"key_type", "key"}
	tagKey = []string{"key_type", "key"}
)

func cat(col string) Ref  { return Ref{Column: col, Table: "base_category", Key: catKey} }
func tags(col string) Ref { return Ref{Column: col, Table: "base_tag", Key: tagKey} }
func prize(col string) Ref {
	return Ref{Column: col, Table: "economy_prize"}
}
func ref(col, table string) Ref { return Ref{Column: col, Table: table} }
func jref(col, path, table string) Ref {
	return Ref{Column: col, Path: path, Table: table}
}
func jtags(col, path string) Ref {
	return Ref{Column: col, Path: path, Table: "base_tag", Key: tagKey}
}

func reqRefs(col, pre string) []Ref {
	return []Ref{
		jref(col, pre+"itemIds[]", "economy_item"),
		jtags(col, pre+"tagIds[]"),
		jtags(col, pre+"excludeTagIds[]"),
		jref(col, pre+"metas[].metaId", "economy_meta"),
		jref(col, pre+"productIds[]", "economy_product"),
	}
}

func BaseAreas() []Area {
	return []Area{
		{Key: "taxonomy", Owner: "base", Order: 10, Title: t2("دسته‌ها، نوع آیتم‌ها و برچسب‌ها", "Categories, item types and tags"), Tables: []Table{
			{Name: "base_category", Key: catKey},
			{Name: "base_tag", Key: tagKey},
		}},
		{Key: "app_versions", Owner: "base", Order: 220, Title: t2("نسخه‌های برنامه", "App versions"), Tables: []Table{
			{Name: "base_app_version", Key: []string{"platform"}, Active: "is_active"},
		}},
		{Key: "server_rules", Owner: "base", Order: 230, Title: t2("قوانین سرور و تنظیمات کنسول", "Server rules and console settings"), Note: t2("حالت تعمیر هم جزو قوانین سرور است؛ پیش از اعمال آن را ببینید.", "Maintenance mode is part of the server rules; check it before you apply."), Tables: []Table{
			{Name: "base_server_rules", Key: []string{"id"}, Refs: []Ref{jtags("value", "maintenance.allowTagIds[]")}},
			{Name: "base_console_settings", Key: []string{"id"}},
			{Name: "base_console_modules", Key: []string{"id"}},
		}},
	}
}

func EconomyAreas() []Area {
	return []Area{
		{Key: "currencies", Owner: "economy", Order: 20, Title: t2("ارزها", "Currencies"), Tables: []Table{
			{Name: "economy_currency", Key: []string{"key"}, Refs: []Ref{cat("category_id"), tags("tag_ids")}},
		}},
		{Key: "metas", Owner: "economy", Order: 30, Title: t2("متاها", "Metas"), Tables: []Table{
			{Name: "economy_meta", Key: []string{"key"}},
		}},
		{Key: "event_types", Owner: "economy", Order: 40, Title: t2("نوع رویدادها و تنظیمات رویداد", "Event types and event settings"), Tables: []Table{
			{Name: "economy_event_type", Key: []string{"key"}, Active: "is_active"},
			{Name: "economy_event_settings", Key: []string{"id"}},
		}},
		{Key: "prizes", Owner: "economy", Order: 50, Title: t2("جایزه‌ها و گروه جایزه", "Prizes and prize groups"), Tables: []Table{
			{Name: "economy_prize", Key: []string{"key"}, Refs: []Ref{cat("category_id"), tags("tag_ids")}},
			{Name: "economy_prize_group", Key: []string{"key"}},
			{Name: "economy_prize_group_member", Key: []string{"group_id", "prize_id"}, Refs: []Ref{ref("group_id", "economy_prize_group"), prize("prize_id")}},
		}},
		{Key: "levels", Owner: "economy", Order: 60, Title: t2("جدول‌های سطح و پله‌ها", "Level tables and steps"), Tables: []Table{
			{Name: "economy_level_table", Key: []string{"key"}, Active: "is_active", Refs: []Ref{ref("meta_id", "economy_meta"), ref("level_meta_id", "economy_meta")}},
			{Name: "economy_level_step", Key: []string{"table_id", "level"}, Refs: []Ref{ref("table_id", "economy_level_table"), prize("prize_id")}},
		}},
		{Key: "items", Owner: "economy", Order: 70, Title: t2("آیتم‌ها", "Items"), Tables: []Table{
			{Name: "economy_item", Key: []string{"key"}, Active: "is_active", Refs: []Ref{cat("type_id"), cat("category_id"), tags("tag_ids"), ref("level_table_id", "economy_level_table")}},
		}},
		{Key: "products", Owner: "economy", Order: 80, Title: t2("محصولات فروشگاه", "Store products"), Tables: []Table{
			{Name: "economy_product", Key: []string{"key"}, Active: "is_active", Refs: []Ref{cat("category_id"), tags("tag_ids"), tags("audience_tag_ids"), prize("prize_id"), ref("level_table_id", "economy_level_table")}},
		}},
		{Key: "event_rules", Owner: "economy", Order: 90, Title: t2("قانون‌های رویداد", "Event rules"), Tables: []Table{
			{Name: "economy_event_rule", Key: []string{"key"}, Active: "is_active", Refs: []Ref{ref("event_type_id", "economy_event_type"), prize("prize_id"), tags("audience_tag_ids")}},
		}},
		{Key: "meta_rules", Owner: "economy", Order: 100, Title: t2("قانون‌ها و رابطه‌های متا", "Meta rules and relations"), Tables: []Table{
			{Name: "economy_meta_rule", Key: []string{"key"}, Active: "is_active", Refs: []Ref{ref("event_type_id", "economy_event_type"), ref("meta_id", "economy_meta"), tags("audience_tag_ids")}},
			{Name: "economy_meta_relation", Key: []string{"key"}, Active: "is_active", Refs: []Ref{ref("meta_id", "economy_meta"), prize("prize_id"), ref("prize_group_id", "economy_prize_group"), tags("audience_tag_ids"), jref("targets", "[].metaId", "economy_meta")}},
		}},
		{Key: "tasks", Owner: "economy", Order: 110, Title: t2("مأموریت‌ها و گروه مأموریت", "Tasks and task groups"), Tables: []Table{
			{Name: "economy_task_group", Key: []string{"key"}, Active: "is_active", Refs: []Ref{cat("category_id"), ref("level_table_id", "economy_level_table"), tags("audience_tag_ids")}},
			{Name: "economy_task", Key: []string{"key"}, Active: "is_active", Refs: []Ref{cat("category_id"), tags("tag_ids"), ref("event_type_id", "economy_event_type"), ref("parent_id", "economy_task"), ref("depends_on", "economy_task"), ref("group_id", "economy_task_group"), prize("prize_id"), tags("audience_tag_ids"), ref("level_table_id", "economy_level_table")}},
		}},
		{Key: "achievements", Owner: "economy", Order: 120, Title: t2("دستاوردها", "Achievements"), Tables: []Table{
			{Name: "economy_achievement", Key: []string{"key"}, Active: "is_active", Refs: []Ref{cat("category_id"), tags("tag_ids"), ref("event_type_id", "economy_event_type"), ref("task_id", "economy_task"), ref("task_group_id", "economy_task_group")}},
			{Name: "economy_achievement_tier", Key: []string{"achievement_id", "tier"}, Refs: []Ref{ref("achievement_id", "economy_achievement"), prize("prize_id")}},
		}},
		{Key: "streaks", Owner: "economy", Order: 130, Title: t2("استریک‌ها", "Streaks"), Tables: []Table{
			{Name: "economy_streak", Key: []string{"key"}, Active: "is_active", Refs: []Ref{ref("event_type_id", "economy_event_type"), prize("infinity_prize_id")}},
			{Name: "economy_streak_reward", Key: []string{"streak_id", "day"}, Refs: []Ref{ref("streak_id", "economy_streak"), prize("prize_id")}},
			{Name: "economy_streak_milestone", Key: []string{"streak_id", "day"}, Refs: []Ref{ref("streak_id", "economy_streak"), prize("prize_id")}},
			{Name: "economy_streak_extra", Key: []string{"streak_id", "day"}, Refs: []Ref{ref("streak_id", "economy_streak"), prize("prize_id")}},
		}},
		{Key: "referral", Owner: "economy", Order: 140, Title: t2("تنظیمات دعوت دوستان", "Referral settings"), Tables: []Table{
			{Name: "economy_referral_settings", Key: []string{"id"}, Refs: []Ref{prize("referrer_prize_id"), prize("referee_prize_id"), ref("goal_event_type_id", "economy_event_type")}},
		}},
		{Key: "offers", Owner: "economy", Order: 150, Title: t2("پیشنهادهای ویژه", "Offers"), Tables: []Table{
			{Name: "economy_offer", Key: []string{"key"}, Active: "is_active", Refs: []Ref{ref("product_id", "economy_product"), tags("audience_tag_ids"), tags("exclude_tag_ids"), ref("level_table_id", "economy_level_table"), jref("meta_conditions", "[].metaId", "economy_meta")}},
		}},
		{Key: "chance_games", Owner: "economy", Order: 160, Title: t2("صندوق‌ها و گردونه‌ها", "Chests and wheels"), Tables: []Table{
			{Name: "economy_chance_game", Key: []string{"key"}, Active: "is_active", Refs: []Ref{ref("level_table_id", "economy_level_table"), jref("prize", "[].prizeId", "economy_prize"), jref("prize", "[].prizeGroupId", "economy_prize_group"), jref("prize", "[].currencyId", "economy_currency"), jref("price", "currencyId", "economy_currency")}},
		}},
		{Key: "tag_discounts", Owner: "economy", Order: 170, Title: t2("تخفیف‌های برچسبی", "Tag discounts"), Tables: []Table{
			{Name: "economy_tag_discount", Key: []string{"key"}, Active: "is_active", Refs: []Ref{tags("tag_ids"), cat("item_type_ids"), ref("product_ids", "economy_product")}},
		}},
		{Key: "games", Owner: "economy", Order: 180, Title: t2("بازی‌های کاتالوگ", "Game catalog"), Tables: []Table{
			{Name: "economy_game", Key: []string{"key"}, Active: "is_active", Refs: append(append([]Ref{cat("category_id"), tags("tag_ids"), jref("rooms", "[].prizeId", "economy_prize"), jref("rooms", "[].prizeGroupId", "economy_prize_group")}, reqRefs("requirements", "")...), reqRefs("rooms", "[].requirements.")...)},
		}},
		{Key: "payment_settings", Owner: "economy", Order: 210, Title: t2("تنظیمات پرداخت", "Payment settings"), Tables: []Table{
			{Name: "economy_payment_settings", Key: []string{"id"}},
		}},
	}
}

func LeaderboardAreas() []Area {
	return []Area{
		{Key: "leaderboards", Owner: "leaderboardx", Order: 190, Title: t2("جدول‌های امتیاز", "Leaderboards"), Note: t2("جدول تازه با اعمال ساخته می‌شود. اگر زمان‌بندی جدولی که بازیکن دارد عوض شود، باید از صفحهٔ خود جدول دوباره شروعش کنید.", "New boards are created on apply. If the schedule of a board with players changes, restart it from the board page."), Tables: []Table{
			{Name: "leaderboardx_board", Key: []string{"key"}, Active: "is_active", Skip: []string{"periods_closed", "finished_at"}},
			{Name: "leaderboardx_settings", Key: []string{"id"}},
		}},
	}
}
