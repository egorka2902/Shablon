// Global state
let currentSlide = 0;
let slideInterval;
let servicesData = [];
let reviewsData = [];
let selectedTime = '';
let isAdminLoggedIn = false;
let adminToken = '';

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    initSlider();
    loadServices();
    loadReviews();
    setMinDate();
});

// Slider
function initSlider() {
    const slides = document.querySelectorAll('.slide');
    const dotsContainer = document.querySelector('.slider-dots');
    
    // Create dots
    slides.forEach((_, i) => {
        const dot = document.createElement('div');
        dot.className = `slider-dot ${i === 0 ? 'active' : ''}`;
        dot.onclick = () => goToSlide(i);
        dotsContainer.appendChild(dot);
    });
    
    // Auto slide
    slideInterval = setInterval(nextSlide, 5000);
}

function goToSlide(index) {
    const slides = document.querySelectorAll('.slide');
    const dots = document.querySelectorAll('.slider-dot');
    
    slides[currentSlide].classList.remove('active');
    dots[currentSlide].classList.remove('active');
    
    currentSlide = index;
    
    slides[currentSlide].classList.add('active');
    dots[currentSlide].classList.add('active');
    
    // Reset auto slide
    clearInterval(slideInterval);
    slideInterval = setInterval(nextSlide, 5000);
}

function nextSlide() {
    const slides = document.querySelectorAll('.slide');
    const next = (currentSlide + 1) % slides.length;
    goToSlide(next);
}

function prevSlide() {
    const slides = document.querySelectorAll('.slide');
    const prev = (currentSlide - 1 + slides.length) % slides.length;
    goToSlide(prev);
}

// Load Services
async function loadServices() {
    try {
        const response = await fetch('/api/services');
        servicesData = await response.json();
        renderServices(servicesData);
        populateServiceSelect();
    } catch (error) {
        console.error('Error loading services:', error);
    }
}

function renderServices(services) {
    const grid = document.getElementById('servicesGrid');
    grid.innerHTML = '';
    
    services.forEach(service => {
        const card = document.createElement('div');
        card.className = 'service-card fade-in';
        card.innerHTML = `
            <div class="service-category">${service.category}</div>
            <div class="service-name">${service.name}</div>
            <div class="service-description">${service.description}</div>
            <div class="service-footer">
                <div class="service-price">${service.price} ₽</div>
                <div class="service-duration">${service.duration}</div>
            </div>
        `;
        grid.appendChild(card);
    });
}

function filterServices(category) {
    // Update active tab
    document.querySelectorAll('.tab').forEach(tab => {
        tab.classList.remove('active');
    });
    event.target.classList.add('active');
    
    // Filter services
    if (category === 'all') {
        renderServices(servicesData);
    } else {
        const filtered = servicesData.filter(s => s.category === category);
        renderServices(filtered);
    }
}

function populateServiceSelect() {
    const select = document.getElementById('bookingService');
    servicesData.forEach(service => {
        const option = document.createElement('option');
        option.value = service.name;
        option.textContent = `${service.name} — ${service.price} ₽ (${service.duration})`;
        select.appendChild(option);
    });
}

// Load Reviews
async function loadReviews() {
    try {
        const response = await fetch('/api/reviews');
        reviewsData = await response.json();
        renderReviews(reviewsData);
    } catch (error) {
        console.error('Error loading reviews:', error);
    }
}

function renderReviews(reviews) {
    const grid = document.getElementById('reviewsGrid');
    grid.innerHTML = '';
    
    reviews.forEach(review => {
        const card = document.createElement('div');
        card.className = 'review-card fade-in';
        
        let stars = '';
        for (let i = 0; i < review.rating; i++) {
            stars += '<span class="star">★</span>';
        }
        
        card.innerHTML = `
            <div class="review-header">
                <span class="review-author">${review.author}</span>
                <span class="review-date">${review.date}</span>
            </div>
            <div class="review-rating">${stars}</div>
            <div class="review-text">${review.text}</div>
        `;
        grid.appendChild(card);
    });
}

// Booking
function setMinDate() {
    const today = new Date().toISOString().split('T')[0];
    document.getElementById('bookingDate').min = today;
}

async function loadTimeSlots() {
    const date = document.getElementById('bookingDate').value;
    if (!date) return;
    
    const container = document.getElementById('timeSlotsContainer');
    const slotsDiv = document.getElementById('timeSlots');
    container.style.display = 'block';
    slotsDiv.innerHTML = '';
    
    // Get existing bookings for this date
    let bookings = [];
    try {
        const response = await fetch('/api/bookings');
        bookings = await response.json();
    } catch (error) {
        console.error('Error loading bookings:', error);
    }
    
    const bookedTimes = bookings
        .filter(b => b.date === date)
        .map(b => b.time);
    
    const timeSlots = [
        '10:00', '10:30', '11:00', '11:30',
        '12:00', '12:30', '13:00', '13:30',
        '14:00', '14:30', '15:00', '15:30',
        '16:00', '16:30', '17:00', '17:30',
        '18:00', '18:30', '19:00', '19:30',
        '20:00', '20:30'
    ];
    
    timeSlots.forEach(time => {
        const slot = document.createElement('div');
        slot.className = 'time-slot';
        slot.textContent = time;
        
        if (bookedTimes.includes(time)) {
            slot.classList.add('booked');
            slot.title = 'Уже занято';
        } else {
            slot.onclick = () => selectTimeSlot(slot, time);
        }
        
        slotsDiv.appendChild(slot);
    });
}

function selectTimeSlot(element, time) {
    document.querySelectorAll('.time-slot').forEach(s => s.classList.remove('selected'));
    element.classList.add('selected');
    selectedTime = time;
}

async function submitBooking(e) {
    e.preventDefault();
    
    if (!selectedTime) {
        alert('Пожалуйста, выберите время');
        return;
    }
    
    const booking = {
        name: document.getElementById('bookingName').value,
        phone: document.getElementById('bookingPhone').value,
        service: document.getElementById('bookingService').value,
        date: document.getElementById('bookingDate').value,
        time: selectedTime,
        notes: document.getElementById('bookingNotes').value
    };
    
    try {
        const response = await fetch('/api/booking', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(booking)
        });
        
        if (response.ok) {
            const data = await response.json();
            alert('✅ Вы успешно записаны!\n\n' +
                  'Имя: ' + data.name + '\n' +
                  'Услуга: ' + data.service + '\n' +
                  'Дата: ' + data.date + '\n' +
                  'Время: ' + data.time);
            closeBookingModal();
            resetBookingForm();
        } else {
            const error = await response.json();
            alert('❌ ' + error.error);
        }
    } catch (error) {
        alert('❌ Ошибка при записи: ' + error.message);
    }
}

function resetBookingForm() {
    document.getElementById('bookingForm').reset();
    selectedTime = '';
    document.getElementById('timeSlotsContainer').style.display = 'none';
    document.querySelectorAll('.time-slot').forEach(s => s.classList.remove('selected'));
}

// Modals
function openBookingModal() {
    document.getElementById('bookingModal').classList.add('active');
    document.body.style.overflow = 'hidden';
}

function closeBookingModal() {
    document.getElementById('bookingModal').classList.remove('active');
    document.body.style.overflow = '';
    resetBookingForm();
}

function openAdminModal() {
    document.getElementById('adminModal').classList.add('active');
    document.body.style.overflow = 'hidden';
    
    if (isAdminLoggedIn) {
        loadBookings();
    }
}

function closeAdminModal() {
    document.getElementById('adminModal').classList.remove('active');
    document.body.style.overflow = '';
}

// Admin
async function adminLogin(e) {
    e.preventDefault();
    
    const username = document.getElementById('adminUsername').value;
    const password = document.getElementById('adminPassword').value;
    
    try {
        const response = await fetch('/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });
        
        if (response.ok) {
            const data = await response.json();
            isAdminLoggedIn = true;
            adminToken = data.token;
            
            document.getElementById('adminLogin').style.display = 'none';
            document.getElementById('adminDashboard').style.display = 'block';
            
            await loadBookings();
        } else {
            alert('❌ Неверный логин или пароль');
        }
    } catch (error) {
        alert('❌ Ошибка авторизации: ' + error.message);
    }
}

function adminLogout() {
    isAdminLoggedIn = false;
    adminToken = '';
    document.getElementById('adminLogin').style.display = 'block';
    document.getElementById('adminDashboard').style.display = 'none';
    document.getElementById('adminUsername').value = '';
    document.getElementById('adminPassword').value = '';
}

async function loadBookings() {
    try {
        const response = await fetch('/api/bookings');
        const bookings = await response.json();
        
        const list = document.getElementById('bookingsList');
        list.innerHTML = '';
        
        if (bookings.length === 0) {
            list.innerHTML = '<p style="text-align:center;color:#999;padding:40px;">Нет записей</p>';
            return;
        }
        
        bookings.forEach(booking => {
            const item = document.createElement('div');
            item.className = 'booking-item fade-in';
            item.innerHTML = `
                <div class="booking-info">
                    <div class="booking-name">${booking.name}</div>
                    <div class="booking-details">
                        <span>📞 ${booking.phone}</span>
                        <span>💅 ${booking.service}</span>
                        <span>📅 ${booking.date}</span>
                        <span>🕐 ${booking.time}</span>
                        ${booking.notes ? `<span>📝 ${booking.notes}</span>` : ''}
                    </div>
                </div>
                <div class="booking-actions">
                    <button class="btn btn-danger" onclick="deleteBooking(${booking.id})">Удалить</button>
                </div>
            `;
            list.appendChild(item);
        });
    } catch (error) {
        console.error('Error loading bookings:', error);
    }
}

async function deleteBooking(id) {
    if (!confirm('Вы уверены, что хотите удалить эту запись?')) return;
    
    try {
        const response = await fetch(`/api/booking/${id}`, {
            method: 'DELETE'
        });
        
        if (response.ok) {
            await loadBookings();
        } else {
            alert('❌ Ошибка при удалении');
        }
    } catch (error) {
        alert('❌ Ошибка: ' + error.message);
    }
}

// Navigation
function scrollToSection(sectionId) {
    const section = document.getElementById(sectionId);
    if (section) {
        section.scrollIntoView({ behavior: 'smooth' });
    }
    
    // Update active nav button
    document.querySelectorAll('.nav-btn').forEach(btn => btn.classList.remove('active'));
    event.target.closest('.nav-btn').classList.add('active');
}

// Close modals on overlay click
document.querySelectorAll('.modal-overlay').forEach(overlay => {
    overlay.addEventListener('click', (e) => {
        if (e.target === overlay) {
            overlay.classList.remove('active');
            document.body.style.overflow = '';
        }
    });
});

// Close modals on Escape key
document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
        document.querySelectorAll('.modal-overlay').forEach(overlay => {
            overlay.classList.remove('active');
        });
        document.body.style.overflow = '';
    }
});
