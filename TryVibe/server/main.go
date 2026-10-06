package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type Booking struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Branch    string `json:"branch"`
	Service   string `json:"service"`
	Date      string `json:"date"`
	Time      string `json:"time"`
	Notes     string `json:"notes"`
	CreatedAt string `json:"created_at"`
}

var bookings []Booking

func init() {
	data, err := os.ReadFile("bookings.json")
	if err == nil {
		json.Unmarshal(data, &bookings)
	}
	if len(bookings) == 0 {
		bookings = []Booking{}
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	http.HandleFunc("/api/services", corsMiddleware(handleServices))
	http.HandleFunc("/api/reviews", corsMiddleware(handleReviews))
	http.HandleFunc("/api/bookings", corsMiddleware(handleBookings))
	http.HandleFunc("/api/booking", corsMiddleware(handleBooking))
	http.HandleFunc("/api/booking/", corsMiddleware(handleBookingByID))
	http.HandleFunc("/api/login", corsMiddleware(handleLogin))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	log.Println("🎨 Rich Nails Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleServices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(getServices())
}

func handleReviews(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(getReviews())
}

func handleBookings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sort.Slice(bookings, func(i, j int) bool {
		if bookings[i].Date == bookings[j].Date {
			return bookings[i].Time < bookings[j].Time
		}
		return bookings[i].Date < bookings[j].Date
	})
	json.NewEncoder(w).Encode(bookings)
}

func handleBooking(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		sort.Slice(bookings, func(i, j int) bool {
			if bookings[i].Date == bookings[j].Date {
				return bookings[i].Time < bookings[j].Time
			}
			return bookings[i].Date < bookings[j].Date
		})
		json.NewEncoder(w).Encode(bookings)

	case "POST":
		var booking Booking
		if err := json.NewDecoder(r.Body).Decode(&booking); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if booking.Name == "" || booking.Phone == "" || booking.Service == "" || booking.Date == "" || booking.Time == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Заполните все обязательные поля"})
			return
		}
		for _, b := range bookings {
			if b.Date == booking.Date && b.Time == booking.Time {
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]string{"error": "Это время уже занято"})
				return
			}
		}
		booking.ID = len(bookings) + 1
		booking.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
		bookings = append(bookings, booking)
		saveBookings()
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(booking)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleBookingByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	url := r.URL.Path
	parts := strings.Split(url, "/")
	idStr := parts[len(parts)-1]

	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID не указан"})
		return
	}

	var id int
	fmt.Sscanf(idStr, "%d", &id)

	switch r.Method {
	case "GET":
		for _, b := range bookings {
			if b.ID == id {
				json.NewEncoder(w).Encode(b)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Запись не найдена"})

	case "DELETE":
		for i, b := range bookings {
			if b.ID == id {
				bookings = append(bookings[:i], bookings[i+1:]...)
				saveBookings()
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"message": "Запись удалена"})
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Запись не найдена"})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if creds.Username == "adminRich" && creds.Password == "nails" {
		token := fmt.Sprintf("%s:%d", creds.Username, time.Now().Unix())
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"token":    token,
			"username": creds.Username,
		})
		return
	}
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": "Неверный логин или пароль"})
}

func saveBookings() {
	data, _ := json.MarshalIndent(bookings, "", "  ")
	os.WriteFile("bookings.json", data, 0644)
}

func getServices() []map[string]interface{} {
	services := []map[string]interface{}{
		{"id": 1, "category": "Маникюр", "name": "Маникюр классический", "description": "Классический обрезной маникюр", "price": 1190, "duration": "40 мин"},
		{"id": 2, "category": "Маникюр", "name": "Маникюр аппаратный", "description": "Аппаратный маникюр без покрытия", "price": 1390, "duration": "45 мин"},
		{"id": 3, "category": "Маникюр", "name": "Маникюр + покрытие гель-лак", "description": "Маникюр с покрытием Rich Nails", "price": 2190, "duration": "90 мин"},
		{"id": 4, "category": "Маникюр", "name": "Маникюр + покрытие Luxio", "description": "Маникюр с покрытием гелем LUXIO", "price": 2890, "duration": "90 мин"},
		{"id": 5, "category": "Маникюр", "name": "Наращивание ногтей 2D", "description": "Наращивание ногтей без снятия", "price": 2627, "duration": "120 мин"},
		{"id": 6, "category": "Маникюр", "name": "Наращивание ногтей", "description": "Полное наращивание ногтей", "price": 3200, "duration": "120 мин"},
		{"id": 7, "category": "Маникюр", "name": "Снятие гель-лака", "description": "Снятие покрытия с рук", "price": 300, "duration": "15 мин"},
		{"id": 8, "category": "Маникюр", "name": "Покрытие гель-лак ADRICOCO", "description": "Покрытие гель-лаком ADRICOCO", "price": 800, "duration": "30 мин"},
		{"id": 9, "category": "Маникюр", "name": "Премиальное покрытие Rich Nails", "description": "Премиальное покрытие гель-лаком", "price": 1100, "duration": "30 мин"},
		{"id": 10, "category": "Маникюр", "name": "Укрепление IBX SYSTEM", "description": "Система укрепления ногтей", "price": 790, "duration": "20 мин"},
		{"id": 11, "category": "Маникюр", "name": "Укрепление IBX BOOST", "description": "Усиленное укрепление IBX", "price": 990, "duration": "20 мин"},
		{"id": 12, "category": "Маникюр", "name": "Лечебное покрытие", "description": "Лечебное покрытие бесцветное", "price": 400, "duration": "15 мин"},
		{"id": 13, "category": "Маникюр", "name": "Парафинотерапия рук", "description": "Уход для восстановления кожи рук", "price": 490, "duration": "30 мин"},
		{"id": 14, "category": "Педикюр", "name": "Педикюр классический", "description": "Классический педикюр", "price": 1790, "duration": "60 мин"},
		{"id": 15, "category": "Педикюр", "name": "Педикюр аппаратный", "description": "Аппаратный педикюр диском", "price": 2190, "duration": "60 мин"},
		{"id": 16, "category": "Педикюр", "name": "Педикюр + покрытие гель-лак", "description": "Педикюр с покрытием Rich Nails", "price": 2542, "duration": "90 мин"},
		{"id": 17, "category": "Педикюр", "name": "Педикюр + покрытие Luxio", "description": "Педикюр с покрытием гелем LUXIO", "price": 3790, "duration": "90 мин"},
		{"id": 18, "category": "Педикюр", "name": "Препаратный педикюр LUXIO", "description": "Препаратный педикюр LUXIO STRADERM", "price": 2390, "duration": "90 мин"},
		{"id": 19, "category": "Педикюр", "name": "Препаратный педикюр KART", "description": "Препаратный педикюр KART (Израиль)", "price": 4090, "duration": "90 мин"},
		{"id": 20, "category": "Педикюр", "name": "Уходовая процедура LUXIO", "description": "Очищение стопы, СПА-уход STRADERM", "price": 690, "duration": "20 мин"},
		{"id": 21, "category": "Педикюр", "name": "Педикюр мужской", "description": "Мужской педикюр", "price": 2490, "duration": "60 мин"},
		{"id": 22, "category": "Педикюр", "name": "Снятие гель-лака с ног", "description": "Снятие покрытия с ног", "price": 300, "duration": "15 мин"},
		{"id": 23, "category": "Брови и ресницы", "name": "Ламинирование ресниц", "description": "Ламинирование ресниц", "price": 2290, "duration": "45 мин"},
		{"id": 24, "category": "Брови и ресницы", "name": "Ламинирование бровей", "description": "Долговременная укладка бровей", "price": 2290, "duration": "40 мин"},
		{"id": 25, "category": "Брови и ресницы", "name": "Ламинирование + ботокс", "description": "Ламинирование с уходом", "price": 2690, "duration": "60 мин"},
		{"id": 26, "category": "Брови и ресницы", "name": "Окрашивание ресниц", "description": "Окрашивание ресниц краской", "price": 600, "duration": "20 мин"},
		{"id": 27, "category": "Брови и ресницы", "name": "Коррекция бровей пинцетом", "description": "Коррекция формы бровей", "price": 690, "duration": "20 мин"},
		{"id": 28, "category": "Брови и ресницы", "name": "Коррекция бровей нитью", "description": "Коррекция формы бровей нитью", "price": 690, "duration": "20 мин"},
		{"id": 29, "category": "Брови и ресницы", "name": "Коррекция бровей сахарной пастой", "description": "Ваксинг бровей", "price": 890, "duration": "25 мин"},
		{"id": 30, "category": "Брови и ресницы", "name": "Коррекция бровей полимером", "description": "Коррекция бровей эластомером", "price": 1190, "duration": "30 мин"},
		{"id": 31, "category": "Брови и ресницы", "name": "Окрашивание бровей краской", "description": "Окрашивание бровей краской", "price": 600, "duration": "15 мин"},
		{"id": 32, "category": "Брови и ресницы", "name": "Окрашивание бровей хной", "description": "Окрашивание бровей хной", "price": 800, "duration": "20 мин"},
		{"id": 33, "category": "Брови и ресницы", "name": "Архитектура бровей", "description": "Архитектура + окрашивание бровей", "price": 1490, "duration": "40 мин"},
		{"id": 34, "category": "Брови и ресницы", "name": "Восстанавливающий уход бровей", "description": "Ботокс для бровей", "price": 500, "duration": "15 мин"},
		{"id": 35, "category": "Брови и ресницы", "name": "Восстанавливающий уход ресниц", "description": "Ботокс для ресниц", "price": 500, "duration": "15 мин"},
		{"id": 36, "category": "Массаж", "name": "Массаж тела", "description": "Общий массаж тела", "price": 1390, "duration": "60 мин"},
		{"id": 37, "category": "Массаж", "name": "Массаж шейно-воротниковой зоны", "description": "Массаж шеи и плеч, 25 мин", "price": 1390, "duration": "25 мин"},
		{"id": 38, "category": "Массаж", "name": "Массаж спины", "description": "Массаж спины, 45 мин", "price": 2590, "duration": "45 мин"},
		{"id": 39, "category": "Массаж", "name": "Массаж рук", "description": "Массаж рук до локтя", "price": 1290, "duration": "30 мин"},
		{"id": 40, "category": "Массаж", "name": "Массаж рук полностью", "description": "Полный массаж рук", "price": 1590, "duration": "45 мин"},
		{"id": 41, "category": "Массаж", "name": "Массаж живота", "description": "Массаж живота", "price": 1290, "duration": "30 мин"},
		{"id": 42, "category": "Массаж", "name": "Массаж поясницы", "description": "Массаж поясницы", "price": 1290, "duration": "30 мин"},
		{"id": 43, "category": "Массаж", "name": "Массаж ягодиц", "description": "Массаж ягодиц", "price": 1490, "duration": "30 мин"},
		{"id": 44, "category": "Депиляция", "name": "Депиляция глубокое бикини", "description": "Сахарная депиляция", "price": 1267, "duration": "30 мин"},
		{"id": 45, "category": "Депиляция", "name": "Депиляция ноги полностью", "description": "Сахарная депиляция ног", "price": 1862, "duration": "45 мин"},
		{"id": 46, "category": "Депиляция", "name": "Депиляция руки", "description": "Сахарная депиляция руки", "price": 1190, "duration": "30 мин"},
		{"id": 47, "category": "Депиляция", "name": "Депиляция подмышек", "description": "Сахарная депиляция подмышек", "price": 890, "duration": "20 мин"},
		{"id": 48, "category": "Депиляция", "name": "Депиляция ноги до колена", "description": "Сахарная депиляция до колена", "price": 1290, "duration": "30 мин"},
		{"id": 49, "category": "Депиляция", "name": "Депиляция ноги выше колена", "description": "Сахарная депиляция выше колена", "price": 1590, "duration": "40 мин"},
		{"id": 50, "category": "Депиляция", "name": "Депиляция пальцы ног", "description": "Сахарная депиляция пальцев", "price": 690, "duration": "15 мин"},
		{"id": 51, "category": "Для мужчин", "name": "Мужской маникюр", "description": "Мужской маникюр", "price": 1390, "duration": "45 мин"},
		{"id": 52, "category": "Для мужчин", "name": "Мужской педикюр", "description": "Мужской педикюр", "price": 2490, "duration": "60 мин"},
		{"id": 53, "category": "Для мужчин", "name": "Мужская стрижка", "description": "Стрижка с мытьем и укладкой", "price": 1390, "duration": "30 мин"},
		{"id": 54, "category": "Для мужчин", "name": "Стрижка машинкой", "description": "Стрижка машинкой, 1 насадка", "price": 890, "duration": "20 мин"},
		{"id": 55, "category": "Для мужчин", "name": "Моделирование бороды", "description": "Формирование формы бороды", "price": 1490, "duration": "20 мин"},
		{"id": 56, "category": "Для мужчин", "name": "Стрижка борода и усы", "description": "Стрижка бороды и усов", "price": 990, "duration": "20 мин"},
		{"id": 57, "category": "Для мужчин", "name": "Стрижка волосы", "description": "Стрижка с пробором/окантовкой", "price": 1690, "duration": "45 мин"},
		{"id": 58, "category": "Для мужчин", "name": "Стрижка + борода", "description": "Комплексная стрижка и борода", "price": 2790, "duration": "60 мин"},
	}
	return services
}

func getReviews() []map[string]interface{} {
	reviews := []map[string]interface{}{
		{"id": 1, "author": "Снежана Жукова", "rating": 5, "text": "Была здесь два раза. Отличный салон, советую. Сделали качественный маникюр с дизайном, наращивание и отличный педикюр. Администрация вежливая и приятная. Спасибо за хорошую работу! ☺️", "date": "12 апреля 2022"},
		{"id": 2, "author": "Анастасия Орлова", "rating": 5, "text": "Айджана — мастер маникюра/педикюра, результатом довольна, быстро, качественно. Экика — косметолог, деликатная чистка, безболезненно, учла пожелания. Спасибо, обращусь ещё! Администратор Маргарита — вежливая, приветливая, заботливая 😊", "date": "13 октября 2025"},
		{"id": 3, "author": "Галина Митрофанова", "rating": 5, "text": "Сегодня была в салоне. Делала маникюр и педикюр. Прихожу уже не первый раз и получаю отличный результат. Все аккуратно, быстро, квалифицированно. Спасибо большое Салехе!", "date": "18 мая 2026"},
		{"id": 4, "author": "Елизавета Михайловская", "rating": 5, "text": "Прекрасное место, мастер своего дела, брови отличные. Рекомендую мастера Эрику. Все быстро, аккуратно и качественно.", "date": "26 февраля 2026"},
		{"id": 5, "author": "Патимат Алиева", "rating": 5, "text": "Хороший салон, была на маникюре у Фирузе, чистая работа и аккуратно. Спасибо администратору Маргарите за кофе. Приду ещё!", "date": "8 октября 2025"},
		{"id": 6, "author": "Ира Александровна", "rating": 5, "text": "Все очень понравилось. Мастер Фируза отлично делает маникюр! Приятная спокойная атмосфера. Хочется возвращаться 🧡", "date": "10 июля 2022"},
		{"id": 7, "author": "Вета Кроликова", "rating": 5, "text": "Понравилась процедура депиляции — быстро и чисто. Удобная запись. Буду приходить ещё!", "date": "10 мая 2025"},
		{"id": 8, "author": "Зухра Наниева", "rating": 5, "text": "Делала коррекцию у Пати, с укреплением, осталась очень довольна, спасибо мастеру.", "date": "15 октября 2025"},
	}
	return reviews
}
