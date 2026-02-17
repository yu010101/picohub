package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

func Seed(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		log.Println("Database already seeded, skipping")
		return nil
	}

	log.Println("Seeding database...")

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	adminHash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 12)

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Users
	_, err = tx.Exec(`INSERT INTO users (username, email, password_hash, display_name, bio, is_admin) VALUES
		('admin', 'admin@picohub.dev', ?, 'PicoHub Admin', 'PicoHub platform administrator', 1),
		('tanaka', 'tanaka@example.com', ?, 'Tanaka Taro', 'IoT enthusiast & PicoClaw developer', 0),
		('suzuki', 'suzuki@example.com', ?, 'Suzuki Hanako', 'Full-stack developer, loves RISC-V', 0)`,
		string(adminHash), string(hash), string(hash))
	if err != nil {
		return err
	}

	// Resolve upload dir for file_path
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "uploads"
	}

	// Skills
	skills := []struct {
		slug, name, desc, version, category string
		authorID                            int
		featured                            int
		tags                                string
		fileHash                            string
	}{
		{"line-messenger", "LINE Messenger", "PicoClawからLINEメッセージを送受信するスキル。Messaging APIを使用してテキスト、画像、スタンプの送信に対応。グループチャットへの通知も可能。", "1.2.0", "messaging", 2, 1, `["line","messaging","chat","notification"]`, "bce6cc8c336bdb39f4d5190093c146c281dc13846f5dec828346e37913d729b6"},
		{"rakuten-shopping", "Rakuten Shopping", "楽天市場の商品検索・価格比較スキル。カテゴリ別検索、価格帯フィルター、ポイント還元率の表示に対応。お買い物マラソン情報も取得可能。", "2.0.1", "shopping", 2, 1, `["rakuten","shopping","price","ecommerce"]`, "c8125f612309a60426ffbbeef78057c3a0477552d8af65e573671fb4c55fc36a"},
		{"weather-reminder", "Weather Reminder", "天気予報を取得してリマインダーを設定するスキル。OpenWeatherMap APIを使用。傘リマインダー、洗濯物アラート、熱中症警告に対応。", "1.0.0", "utility", 3, 1, `["weather","reminder","alert","openweathermap"]`, "9ee3c92243eb16bc04fcc198e977d8d77abab287a8714857a21c4eba94ad4c40"},
		{"mercari-lister", "Mercari Lister", "メルカリ出品テキストを自動生成するスキル。商品写真から説明文を作成し、適切なカテゴリ・価格を提案。テンプレート機能付き。", "1.1.0", "commerce", 3, 0, `["mercari","listing","commerce","automation"]`, "a0a4dd8b10f6e86f763d9399ff6b8064b812a1bcd42df7da25b16f91fe2cc4d8"},
		{"notion-lite", "Notion Lite", "軽量Notion連携スキル。ページの作成・読取・更新に対応。データベースへのレコード追加、日報テンプレートの自動生成が可能。", "0.9.0", "productivity", 2, 1, `["notion","productivity","notes","database"]`, "fd4e1fe49434e976e277f192b7abeacef1c6e921cc409cae7dfa3505bff15f3b"},
	}

	for _, s := range skills {
		filePath := filepath.Join(uploadDir, s.slug+".zip")
		_, err = tx.Exec(`INSERT INTO skills (slug, name, description, version, category, author_id, file_path, file_hash, scan_status, is_featured, tags, download_count) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'clean', ?, ?, ?)`,
			s.slug, s.name, s.desc, s.version, s.category, s.authorID, filePath, s.fileHash, s.featured, s.tags,
			100+len(s.slug)*17)
		if err != nil {
			return err
		}
	}

	// Reviews
	reviews := []struct {
		skillSlug string
		userID    int
		rating    int
		title     string
		body      string
	}{
		{"line-messenger", 3, 5, "完璧なLINE連携", "PicoClawからLINEに通知を送れるようになって非常に便利。セットアップも簡単でした。"},
		{"line-messenger", 1, 4, "便利だが改善の余地あり", "基本機能は十分。リッチメニュー対応があるとさらに良い。"},
		{"rakuten-shopping", 3, 5, "お買い物が楽になった", "価格比較が自動でできるのは素晴らしい。ポイント計算も正確。"},
		{"rakuten-shopping", 1, 4, "API制限に注意", "機能は充実しているが、API呼び出し回数の制限に注意が必要。"},
		{"weather-reminder", 2, 5, "毎朝助かっています", "傘リマインダーが特に便利。雨の日に傘を忘れなくなりました。"},
		{"mercari-lister", 2, 4, "出品が楽になった", "説明文の自動生成は便利。写真からの価格提案精度がもう少し上がると嬉しい。"},
		{"notion-lite", 3, 5, "日報作成が自動化できた", "毎日の日報テンプレートが自動生成されるので手間が省ける。軽量で速い。"},
		{"notion-lite", 1, 4, "シンプルで使いやすい", "Notion APIの複雑さをうまく隠蔽している。ブロック操作にも対応してほしい。"},
	}

	for _, r := range reviews {
		_, err = tx.Exec(`INSERT INTO reviews (skill_id, user_id, rating, title, body)
			VALUES ((SELECT id FROM skills WHERE slug = ?), ?, ?, ?, ?)`,
			r.skillSlug, r.userID, r.rating, r.title, r.body)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Println("Database seeded successfully")
	return nil
}
