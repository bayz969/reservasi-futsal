// FutsalBook SPA Client Core Logic

const API_BASE = '/api/v1';

// Global toast & loader helper
function showToast(message, typeClass = 'bg-primary') {
  const toastEl = document.getElementById('appToast');
  const toastMsg = document.getElementById('toastMessage');
  
  toastEl.className = `toast align-items-center text-white border-0 ${typeClass}`;
  toastMsg.innerText = message;
  
  const toast = new bootstrap.Toast(toastEl);
  toast.show();
}

function showLoader() {
  document.getElementById('loadingOverlay').classList.remove('d-none');
  document.getElementById('loadingOverlay').classList.add('d-flex');
}

function hideLoader() {
  document.getElementById('loadingOverlay').classList.remove('d-flex');
  document.getElementById('loadingOverlay').classList.add('d-none');
}

// Global API Fetch helper
async function apiCall(endpoint, method = 'GET', body = null) {
  const token = localStorage.getItem('token');
  const options = {
    method,
    headers: {
      'Content-Type': 'application/json'
    }
  };
  
  if (token) {
    options.headers['Authorization'] = `Bearer ${token}`;
  }
  
  if (body) {
    options.body = JSON.stringify(body);
  }
  
  showLoader();
  try {
    const res = await fetch(`${API_BASE}${endpoint}`, options);
    let data = {};
    const contentType = res.headers.get("content-type");
    if (contentType && contentType.includes("application/json")) {
      data = await res.json();
    }
    
    hideLoader();
    
    if (res.status === 401) {
      localStorage.clear();
      window.location.hash = '#/login';
      showToast(data.error || 'Sesi telah berakhir. Silakan login kembali.', 'bg-danger');
      throw new Error('Unauthorized');
    }
    
    if (!res.ok) {
      throw new Error(data.error || 'Terjadi kesalahan sistem');
    }
    
    return data;
  } catch (err) {
    hideLoader();
    if (err.message !== 'Unauthorized') {
      showToast(err.message, 'bg-danger');
    }
    throw err;
  }
}

// Router Event Dispatcher (called by router.js)
function triggerPageInit(path, params) {
  if (path === '/login') {
    initLogin();
  } else if (path === '/register') {
    initRegister();
  } else if (path === '/forgot-password') {
    initForgotPassword();
  } else if (path === '/reset-password') {
    initResetPassword();
  } else if (path === '/dashboard') {
    loadUserDashboard();
  } else if (path === '/fields') {
    loadUserFields();
  } else if (path.startsWith('/fields/')) {
    loadFieldDetail(params.id);
  } else if (path.startsWith('/booking/')) {
    loadBookingForm(params.field_id);
  } else if (path === '/bookings') {
    loadUserBookings();
  } else if (path === '/profile') {
    loadProfile();
  } else if (path === '/admin/dashboard') {
    loadAdminDashboard();
  } else if (path === '/admin/fields') {
    loadAdminFields();
  } else if (path === '/admin/fields/new') {
    initCreateField();
  } else if (path.startsWith('/admin/fields/edit/')) {
    loadFieldEdit(params.id);
  } else if (path === '/admin/bookings') {
    loadAdminBookings();
  } else if (path === '/admin/users') {
    loadAdminUsers();
  } else if (path === '/admin/reports') {
    loadAdminReports();
  }
}

// === AUTHENTICATION LOGIC ===

function initLogin() {
  const form = document.getElementById('form-login');
  form.onsubmit = async (e) => {
    e.preventDefault();
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;
    
    try {
      const data = await apiCall('/auth/login', 'POST', { email, password });
      localStorage.setItem('token', data.token);
      localStorage.setItem('user', JSON.stringify(data.user));
      
      showToast(`Selamat datang kembali, ${data.user.nama}!`, 'bg-success');
      
      // Redirect based on role
      window.location.hash = data.user.role === 'admin' ? '#/admin/dashboard' : '#/dashboard';
    } catch (err) {
      // Toast already handled by apiCall
    }
  };
}

function initRegister() {
  const form = document.getElementById('form-register');
  form.onsubmit = async (e) => {
    e.preventDefault();
    const nama = document.getElementById('register-nama').value;
    const email = document.getElementById('register-email').value;
    const nomor_hp = document.getElementById('register-hp').value;
    const password = document.getElementById('register-password').value;
    
    try {
      const data = await apiCall('/auth/register', 'POST', { nama, email, nomor_hp, password });
      showToast(data.message, 'bg-success');
      window.location.hash = '#/login';
    } catch (err) {}
  };
}

function initForgotPassword() {
  const form = document.getElementById('form-forgot-password');
  form.onsubmit = async (e) => {
    e.preventDefault();
    const email = document.getElementById('forgot-email').value;
    
    try {
      const data = await apiCall('/auth/forgot-password', 'POST', { email });
      showToast(data.message, 'bg-success');
      
      // Auto-populate simulation token in URL to make testing easy
      if (data.token_simulation) {
        setTimeout(() => {
          if (confirm(`[DEBUG] Simulasi Token Reset ditemukan:\n\n${data.token_simulation}\n\nIngin langsung membuka halaman Reset Password?`)) {
            window.location.hash = `#/reset-password?token=${data.token_simulation}`;
          }
        }, 1000);
      }
    } catch (err) {}
  };
}

function initResetPassword() {
  // Parse token from hash parameter if present
  // Format: #/reset-password?token=XYZ
  const hash = window.location.hash;
  let token = '';
  if (hash.includes('?')) {
    const query = hash.split('?')[1];
    const params = new URLSearchParams(query);
    token = params.get('token') || '';
  }
  
  document.getElementById('reset-token-input').value = token;
  
  if (!token) {
    showToast('Token reset password tidak ditemukan di URL', 'bg-warning');
  }

  const form = document.getElementById('form-reset-password');
  form.onsubmit = async (e) => {
    e.preventDefault();
    const currentToken = document.getElementById('reset-token-input').value;
    const newPassword = document.getElementById('reset-new-password').value;
    
    try {
      const data = await apiCall('/auth/reset-password', 'POST', { token: currentToken, new_password: newPassword });
      showToast(data.message, 'bg-success');
      window.location.hash = '#/login';
    } catch (err) {}
  };
}

async function handleLogout() {
  if (confirm('Apakah Anda yakin ingin keluar?')) {
    try {
      await apiCall('/auth/logout', 'POST');
    } catch (e) {}
    localStorage.clear();
    window.location.hash = '#/login';
    showToast('Anda berhasil logout.', 'bg-success');
  }
}

// === USER PORTAL LOGIC ===

async function loadUserDashboard() {
  try {
    const data = await apiCall('/dashboard/user');
    
    // Set greeting
    const user = JSON.parse(localStorage.getItem('user'));
    document.getElementById('dash-greeting').innerText = `Halo, ${user.nama}!`;
    
    // Stats
    document.getElementById('dash-active-count').innerText = data.active_bookings_count;
    document.getElementById('dash-next-field').innerText = data.next_booking_field;
    
    if (data.next_booking_date !== '-') {
      document.getElementById('dash-next-time').innerText = `${formatDate(data.next_booking_date)} pukul ${data.next_booking_time.substring(0, 5)}`;
    } else {
      document.getElementById('dash-next-time').innerText = '-';
    }
    
    // Table
    const tbody = document.getElementById('dash-recent-bookings-list');
    tbody.innerHTML = '';
    
    if (data.recent_bookings.length === 0) {
      tbody.innerHTML = `<tr><td colspan="6" class="text-center text-muted">Belum ada aktivitas booking.</td></tr>`;
      return;
    }
    
    data.recent_bookings.forEach(b => {
      const tr = document.createElement('tr');
      tr.className = 'row-item';
      tr.innerHTML = `
        <td class="fw-semibold">${b.field_nama}</td>
        <td>${formatDate(b.tanggal)}</td>
        <td>${b.jam_mulai.substring(0,5)} - ${b.jam_selesai.substring(0,5)}</td>
        <td>${formatRupiah(b.total_harga)}</td>
        <td><span class="badge-status badge-${b.status}">${b.status}</span></td>
        <td>
          <button class="btn btn-sm btn-outline-info" onclick="viewBookingDetail(${b.id})"><i class="fa-solid fa-eye"></i> Detail</button>
        </td>
      `;
      tbody.appendChild(tr);
    });
  } catch (err) {}
}

async function loadUserFields() {
  const form = document.getElementById('form-filter-fields');
  
  // Render fields helper
  const renderFields = async () => {
    const search = document.getElementById('filter-search').value;
    const lokasi = document.getElementById('filter-lokasi').value;
    const maxHarga = document.getElementById('filter-harga').value;
    
    let query = `?status=tersedia`;
    if (search) query += `&search=${encodeURIComponent(search)}`;
    if (lokasi) query += `&lokasi=${encodeURIComponent(lokasi)}`;
    if (maxHarga) query += `&harga_max=${maxHarga}`;
    
    try {
      const fields = await apiCall(`/fields${query}`);
      const grid = document.getElementById('fields-grid-list');
      grid.innerHTML = '';
      
      if (fields.length === 0) {
        grid.innerHTML = `<div class="col-12 text-center text-muted my-5"><i class="fa-solid fa-face-frown display-4 mb-3"></i><h5>Tidak ada lapangan yang sesuai dengan kriteria filter.</h5></div>`;
        return;
      }
      
      fields.forEach(f => {
        const col = document.createElement('div');
        col.className = 'col-md-6 col-lg-4';
        col.innerHTML = `
          <div class="glass-card h-100 d-flex flex-column" style="overflow:hidden;">
            <div class="field-img-container">
              <img src="${f.foto}" class="field-img" alt="${f.nama_lapangan}">
              <span class="field-price-tag">${formatRupiah(f.harga)}/Jam</span>
            </div>
            <div class="p-4 d-flex flex-column flex-grow-1">
              <h5 class="fw-bold mb-1 text-white">${f.nama_lapangan}</h5>
              <p class="text-primary small mb-3"><i class="fa-solid fa-location-dot me-1"></i> ${f.lokasi}</p>
              <p class="text-muted small text-truncate-3 flex-grow-1">${f.deskripsi}</p>
              <div class="d-flex gap-2 mt-3">
                <a href="#/fields/${f.id}" class="btn btn-outline-light w-50 py-2 btn-animate text-center small"><i class="fa-solid fa-circle-info me-1"></i> Rincian</a>
                <a href="#/booking/${f.id}" class="btn btn-primary w-50 py-2 btn-animate text-center small fw-semibold"><i class="fa-solid fa-calendar-check me-1"></i> Booking</a>
              </div>
            </div>
          </div>
        `;
        grid.appendChild(col);
      });
    } catch (err) {}
  };

  // Bind submit
  form.onsubmit = (e) => {
    e.preventDefault();
    renderFields();
  };

  // Initial load
  renderFields();
}

async function loadFieldDetail(id) {
  try {
    const f = await apiCall(`/fields/${id}`);
    
    document.getElementById('field-detail-img').src = f.foto;
    document.getElementById('field-detail-name').innerText = f.nama_lapangan;
    document.getElementById('field-detail-location').innerHTML = `<i class="fa-solid fa-location-dot me-2"></i>${f.lokasi}`;
    document.getElementById('field-detail-price').innerText = `${formatRupiah(f.harga)} / Jam`;
    document.getElementById('field-detail-description').innerText = f.deskripsi;
    
    // Status Badge
    const statusBadge = document.getElementById('field-detail-status');
    statusBadge.innerText = f.status === 'tersedia' ? 'Tersedia' : 'Tidak Tersedia';
    statusBadge.className = f.status === 'tersedia' ? 'badge bg-success mb-2' : 'badge bg-danger mb-2';

    // Facilities
    const facContainer = document.getElementById('field-detail-facilities');
    facContainer.innerHTML = '';
    
    const list = f.fasilitas.split(',');
    list.forEach(item => {
      if (item.trim() !== "") {
        const badge = document.createElement('span');
        badge.className = 'badge bg-secondary border border-secondary';
        badge.innerText = item.trim();
        facContainer.appendChild(badge);
      }
    });

    // Booking Button
    const bookingBtn = document.getElementById('field-detail-booking-btn');
    if (f.status === 'tersedia') {
      bookingBtn.href = `#/booking/${f.id}`;
      bookingBtn.className = 'btn btn-primary w-100 py-3 btn-animate fw-bold';
    } else {
      bookingBtn.removeAttribute('href');
      bookingBtn.className = 'btn btn-secondary w-100 py-3 disabled fw-bold';
    }

  } catch (err) {}
}

async function loadBookingForm(fieldId) {
  try {
    const f = await apiCall(`/fields/${fieldId}`);
    
    // Populate Info Card
    document.getElementById('booking-field-img').src = f.foto;
    document.getElementById('booking-field-name').innerText = f.nama_lapangan;
    document.getElementById('booking-field-location').innerHTML = `<i class="fa-solid fa-location-dot me-1"></i>${f.lokasi}`;
    document.getElementById('booking-field-price').innerText = `${formatRupiah(f.harga)} / jam`;
    
    // Form setup
    document.getElementById('booking-field-id').value = f.id;
    document.getElementById('booking-tanggal').min = new Date().toISOString().split('T')[0];
    
    // Calculation Helper
    const calcTotal = () => {
      const durasi = parseInt(document.getElementById('booking-durasi').value) || 0;
      const total = f.harga * durasi;
      document.getElementById('booking-calc-price').innerText = formatRupiah(total);
    };

    document.getElementById('booking-durasi').oninput = calcTotal;
    calcTotal(); // initial calc

    // Submit handler
    const form = document.getElementById('form-create-booking');
    form.onsubmit = async (e) => {
      e.preventDefault();
      
      const tanggal = document.getElementById('booking-tanggal').value;
      const jam_mulai = document.getElementById('booking-jam').value;
      const durasi = parseInt(document.getElementById('booking-durasi').value);
      const catatan = document.getElementById('booking-catatan').value;
      
      const payload = {
        field_id: f.id,
        tanggal,
        jam_mulai,
        durasi,
        catatan
      };
      
      try {
        const data = await apiCall('/bookings', 'POST', payload);
        showToast(data.message, 'bg-success');
        window.location.hash = '#/bookings';
      } catch (err) {}
    };

  } catch (err) {}
}

async function loadUserBookings() {
  try {
    const bookings = await apiCall('/bookings');
    const tbody = document.getElementById('user-bookings-list');
    tbody.innerHTML = '';

    if (bookings.length === 0) {
      tbody.innerHTML = `<tr><td colspan="7" class="text-center text-muted py-5">Anda belum memiliki riwayat reservasi.</td></tr>`;
      return;
    }

    bookings.forEach(b => {
      const tr = document.createElement('tr');
      tr.className = 'row-item';
      
      // Actions: show Cancel button ONLY if status is Pending
      let actionBtn = `<button class="btn btn-sm btn-outline-info me-1" onclick="viewBookingDetail(${b.id})"><i class="fa-solid fa-eye"></i> Detail</button>`;
      if (b.status === 'pending') {
        actionBtn += `<button class="btn btn-sm btn-danger" onclick="cancelBooking(${b.id})"><i class="fa-solid fa-trash"></i> Batal</button>`;
      }

      tr.innerHTML = `
        <td>#${b.id}</td>
        <td class="fw-bold text-white">${b.field_nama}</td>
        <td>${formatDate(b.tanggal)}</td>
        <td>${b.jam_mulai.substring(0,5)} - ${b.jam_selesai.substring(0,5)}</td>
        <td>${formatRupiah(b.total_harga)}</td>
        <td><span class="badge-status badge-${b.status}">${b.status}</span></td>
        <td>${actionBtn}</td>
      `;
      tbody.appendChild(tr);
    });
  } catch (err) {}
}

async function loadProfile() {
  try {
    const user = await apiCall('/users/me');
    document.getElementById('profile-email').value = user.email;
    document.getElementById('profile-nama').value = user.nama;
    document.getElementById('profile-hp').value = user.nomor_hp;
    
    const form = document.getElementById('form-update-profile');
    form.onsubmit = async (e) => {
      e.preventDefault();
      
      const nama = document.getElementById('profile-nama').value;
      const nomor_hp = document.getElementById('profile-hp').value;
      const password = document.getElementById('profile-password').value;
      
      const payload = { nama, nomor_hp };
      if (password) {
        payload.password = password;
      }
      
      try {
        const data = await apiCall('/users/me', 'PUT', payload);
        showToast(data.message, 'bg-success');
        
        // Update local session storage
        localStorage.setItem('user', JSON.stringify(data.user));
        
        // Refresh sidebar view
        document.getElementById('sidebar-username').innerText = data.user.nama;
        document.getElementById('sidebar-avatar').innerText = data.user.nama.charAt(0).toUpperCase();
        
        // Clear password box
        document.getElementById('profile-password').value = '';
      } catch (err) {}
    };
  } catch (err) {}
}

async function viewBookingDetail(id) {
  try {
    const b = await apiCall(`/bookings/${id}`);
    const modalContent = document.getElementById('bookingDetailContent');
    
    modalContent.innerHTML = `
      <div class="mb-3 text-center">
        <span class="badge-status badge-${b.status} fs-6 mb-2">${b.status}</span>
        <h4 class="fw-bold mb-1">${b.field_nama}</h4>
        <p class="text-muted small"><i class="fa-solid fa-location-dot me-1"></i> ${b.field_lokasi}</p>
      </div>
      
      <div class="row g-2 border-top border-secondary border-opacity-10 pt-3">
        <div class="col-6 text-muted">Pemesan:</div>
        <div class="col-6 text-end fw-semibold">${b.user_nama}</div>
        
        <div class="col-6 text-muted">No. HP:</div>
        <div class="col-6 text-end">${b.user_nomor_hp}</div>

        <div class="col-6 text-muted">Tanggal:</div>
        <div class="col-6 text-end fw-semibold">${formatDate(b.tanggal)}</div>

        <div class="col-6 text-muted">Waktu Main:</div>
        <div class="col-6 text-end fw-semibold text-warning">${b.jam_mulai.substring(0,5)} s.d ${b.jam_selesai.substring(0,5)}</div>

        <div class="col-6 text-muted">Durasi:</div>
        <div class="col-6 text-end">${calculateHourDiff(b.jam_mulai, b.jam_selesai)} Jam</div>

        <div class="col-6 text-muted">Biaya Sewa:</div>
        <div class="col-6 text-end fw-bold text-info">${formatRupiah(b.total_harga)}</div>
      </div>
      
      <div class="mt-3 p-3 bg-black bg-opacity-25 rounded-3 border border-secondary border-opacity-10">
        <span class="text-muted small d-block mb-1">Catatan Tambahan:</span>
        <p class="mb-0 small text-white">${b.catatan || 'Tidak ada catatan khusus.'}</p>
      </div>
    `;
    
    const modal = new bootstrap.Modal(document.getElementById('bookingDetailModal'));
    modal.show();
  } catch (err) {}
}

async function cancelBooking(id) {
  if (confirm('Apakah Anda yakin ingin membatalkan dan menghapus booking ini?')) {
    try {
      const data = await apiCall(`/bookings/${id}`, 'DELETE');
      showToast(data.message, 'bg-success');
      loadUserBookings();
    } catch (err) {}
  }
}


// === ADMIN PORTAL LOGIC ===

async function loadAdminDashboard() {
  try {
    const data = await apiCall('/dashboard/admin');
    
    document.getElementById('admin-dash-users').innerText = data.total_users;
    document.getElementById('admin-dash-fields').innerText = data.total_fields;
    document.getElementById('admin-dash-bookings').innerText = data.total_bookings;
    document.getElementById('admin-dash-pending').innerText = data.pending_bookings;
    document.getElementById('admin-dash-today').innerText = data.bookings_today;
  } catch (err) {}
}

async function loadAdminFields() {
  try {
    const fields = await apiCall('/fields');
    const tbody = document.getElementById('admin-fields-list');
    tbody.innerHTML = '';
    
    if (fields.length === 0) {
      tbody.innerHTML = `<tr><td colspan="6" class="text-center text-muted py-5">Belum ada lapangan terdaftar.</td></tr>`;
      return;
    }
    
    fields.forEach(f => {
      const tr = document.createElement('tr');
      tr.className = 'row-item';
      
      tr.innerHTML = `
        <td><img src="${f.foto}" class="rounded-3" style="width: 70px; height: 50px; object-fit: cover;" alt="Mini Foto"></td>
        <td class="fw-bold text-white">${f.nama_lapangan}</td>
        <td>${f.lokasi}</td>
        <td>${formatRupiah(f.harga)}</td>
        <td>
          <span class="badge ${f.status === 'tersedia' ? 'bg-success' : 'bg-danger'}">${f.status === 'tersedia' ? 'Tersedia' : 'Non-Aktif'}</span>
        </td>
        <td>
          <a href="#/admin/fields/edit/${f.id}" class="btn btn-sm btn-outline-warning me-1"><i class="fa-solid fa-pen"></i> Edit</a>
          <button class="btn btn-sm btn-danger" onclick="deleteField(${f.id})"><i class="fa-solid fa-trash"></i> Hapus</button>
        </td>
      `;
      tbody.appendChild(tr);
    });
  } catch (err) {}
}

function initCreateField() {
  const form = document.getElementById('form-create-field');
  
  // Clear previous values
  document.getElementById('new-field-name').value = '';
  document.getElementById('new-field-location').value = '';
  document.getElementById('new-field-price').value = '';
  document.getElementById('new-field-facilities').value = '';
  document.getElementById('new-field-description').value = '';
  document.getElementById('new-field-foto').value = '';

  form.onsubmit = async (e) => {
    e.preventDefault();
    const nama_lapangan = document.getElementById('new-field-name').value;
    const lokasi = document.getElementById('new-field-location').value;
    const harga = parseFloat(document.getElementById('new-field-price').value);
    const status = document.getElementById('new-field-status').value;
    const fasilitas = document.getElementById('new-field-facilities').value;
    const deskripsi = document.getElementById('new-field-description').value;
    const foto = document.getElementById('new-field-foto').value;
    
    const payload = { nama_lapangan, lokasi, harga, status, fasilitas, deskripsi, foto };
    
    try {
      const data = await apiCall('/fields', 'POST', payload);
      showToast(data.message, 'bg-success');
      window.location.hash = '#/admin/fields';
    } catch (err) {}
  };
}

async function loadFieldEdit(id) {
  try {
    const f = await apiCall(`/fields/${id}`);
    
    document.getElementById('edit-field-id').value = f.id;
    document.getElementById('edit-field-name').value = f.nama_lapangan;
    document.getElementById('edit-field-location').value = f.lokasi;
    document.getElementById('edit-field-price').value = f.harga;
    document.getElementById('edit-field-status').value = f.status;
    document.getElementById('edit-field-facilities').value = f.fasilitas;
    document.getElementById('edit-field-description').value = f.deskripsi;
    document.getElementById('edit-field-foto').value = f.foto;
    
    const form = document.getElementById('form-edit-field');
    form.onsubmit = async (e) => {
      e.preventDefault();
      
      const nama_lapangan = document.getElementById('edit-field-name').value;
      const lokasi = document.getElementById('edit-field-location').value;
      const harga = parseFloat(document.getElementById('edit-field-price').value);
      const status = document.getElementById('edit-field-status').value;
      const fasilitas = document.getElementById('edit-field-facilities').value;
      const deskripsi = document.getElementById('edit-field-description').value;
      const foto = document.getElementById('edit-field-foto').value;
      
      const payload = { nama_lapangan, lokasi, harga, status, fasilitas, deskripsi, foto };
      
      try {
        const data = await apiCall(`/fields/${id}`, 'PUT', payload);
        showToast(data.message, 'bg-success');
        window.location.hash = '#/admin/fields';
      } catch (err) {}
    };

  } catch (err) {}
}

async function deleteField(id) {
  // Confirmation Alert dialog (FR-21)
  if (confirm('Apakah Anda yakin ingin menghapus data lapangan ini? Lapangan yang memiliki booking aktif tidak dapat dihapus.')) {
    try {
      const data = await apiCall(`/fields/${id}`, 'DELETE');
      showToast(data.message, 'bg-success');
      loadAdminFields();
    } catch (err) {}
  }
}

async function loadAdminBookings() {
  try {
    const bookings = await apiCall('/bookings');
    const tbody = document.getElementById('admin-bookings-list');
    tbody.innerHTML = '';
    
    if (bookings.length === 0) {
      tbody.innerHTML = `<tr><td colspan="8" class="text-center text-muted py-5">Belum ada booking masuk.</td></tr>`;
      return;
    }
    
    bookings.forEach(b => {
      const tr = document.createElement('tr');
      tr.className = 'row-item';
      
      // Approve/Reject/Finished logic
      let actions = '';
      if (b.status === 'pending') {
        actions = `
          <button class="btn btn-sm btn-success me-1 btn-animate" onclick="updateBookingStatus(${b.id}, 'approved')"><i class="fa-solid fa-check"></i> Setujui</button>
          <button class="btn btn-sm btn-danger btn-animate" onclick="updateBookingStatus(${b.id}, 'rejected')"><i class="fa-solid fa-xmark"></i> Tolak</button>
        `;
      } else if (b.status === 'approved') {
        actions = `
          <button class="btn btn-sm btn-info text-white btn-animate" onclick="updateBookingStatus(${b.id}, 'finished')"><i class="fa-solid fa-flag-checkered"></i> Selesai</button>
        `;
      } else {
        actions = `<span class="text-muted small">-</span>`;
      }
      
      tr.innerHTML = `
        <td>#${b.id}</td>
        <td>
          <div class="fw-semibold text-white">${b.user_nama}</div>
          <span class="text-muted small">${b.user_nomor_hp}</span>
        </td>
        <td class="fw-bold">${b.field_nama}</td>
        <td>
          <div class="small">${formatDate(b.tanggal)}</div>
          <span class="text-warning small">${b.jam_mulai.substring(0,5)} - ${b.jam_selesai.substring(0,5)}</span>
        </td>
        <td>${formatRupiah(b.total_harga)}</td>
        <td><span class="badge-status badge-${b.status}">${b.status}</span></td>
        <td><span class="small text-truncate-2 d-inline-block" style="max-width:120px;" title="${b.catatan}">${b.catatan || '-'}</span></td>
        <td>${actions}</td>
      `;
      tbody.appendChild(tr);
    });
  } catch (err) {}
}

async function updateBookingStatus(id, newStatus) {
  // Confirm actions
  let confMsg = `Setujui booking #${id}?`;
  if (newStatus === 'rejected') confMsg = `Tolak booking #${id}?`;
  if (newStatus === 'finished') confMsg = `Ubah status booking #${id} menjadi Selesai?`;
  
  if (confirm(confMsg)) {
    try {
      const data = await apiCall(`/bookings/${id}/status`, 'PUT', { status: newStatus });
      showToast(data.message, 'bg-success');
      loadAdminBookings();
    } catch (err) {}
  }
}

async function loadAdminUsers() {
  try {
    const users = await apiCall('/users');
    const tbody = document.getElementById('admin-users-list');
    tbody.innerHTML = '';
    
    users.forEach(u => {
      const tr = document.createElement('tr');
      tr.className = 'row-item';
      
      let toggleBtn = '';
      if (u.role === 'admin') {
        toggleBtn = `<span class="text-muted small">-</span>`;
      } else {
        const isAct = u.status === 'active';
        toggleBtn = `
          <button class="btn btn-sm ${isAct ? 'btn-outline-danger' : 'btn-outline-success'}" onclick="toggleUserStatus(${u.id}, '${isAct ? 'inactive' : 'active'}')">
            <i class="fa-solid ${isAct ? 'fa-user-slash' : 'fa-user-check'}"></i> ${isAct ? 'Nonaktifkan' : 'Aktifkan'}
          </button>
        `;
      }
      
      tr.innerHTML = `
        <td>#${u.id}</td>
        <td class="fw-semibold text-white">${u.nama}</td>
        <td>${u.email}</td>
        <td>${u.nomor_hp}</td>
        <td><span class="badge ${u.role === 'admin' ? 'bg-primary' : 'bg-secondary'}">${u.role}</span></td>
        <td>
          <span class="badge ${u.status === 'active' ? 'bg-success' : 'bg-danger'}">${u.status === 'active' ? 'Aktif' : 'Non-Aktif'}</span>
        </td>
        <td>${toggleBtn}</td>
      `;
      tbody.appendChild(tr);
    });
  } catch (err) {}
}

async function toggleUserStatus(id, newStatus) {
  const label = newStatus === 'active' ? 'mengaktifkan' : 'menonaktifkan';
  if (confirm(`Apakah Anda yakin ingin ${label} akun user ini?`)) {
    try {
      const data = await apiCall(`/users/${id}/status`, 'PUT', { status: newStatus });
      showToast(data.message, 'bg-success');
      loadAdminUsers();
    } catch (err) {}
  }
}

async function loadAdminReports() {
  try {
    // 1. Fetch booking report (revenue & monthly statistics)
    const bookRep = await apiCall('/reports/bookings');
    document.getElementById('report-total-revenue').innerText = formatRupiah(bookRep.total_revenue);
    
    // Render custom bar charts (with pure css/html layout defined in stylesheet)
    const chartContainer = document.getElementById('report-monthly-chart-bars');
    chartContainer.innerHTML = '';
    
    const monthlyList = bookRep.monthly_reports || [];
    
    if (monthlyList.length === 0) {
      chartContainer.innerHTML = `<span class="text-muted py-5 text-center w-100">Belum ada data transaksi bulanan.</span>`;
    } else {
      // Find max count to scale heights
      let maxCount = 1;
      monthlyList.forEach(m => {
        if (m.count > maxCount) maxCount = m.count;
      });
      
      monthlyList.forEach(m => {
        const heightPercent = Math.max(10, (m.count / maxCount) * 100);
        
        const bar = document.createElement('div');
        bar.className = 'chart-bar-item';
        bar.style.height = `${heightPercent}%`;
        bar.title = `${m.month}: ${m.count} Booking`;
        
        bar.innerHTML = `
          <span class="chart-bar-val text-white fw-bold">${m.count}</span>
          <span class="chart-bar-label">${formatMonthString(m.month)}</span>
        `;
        chartContainer.appendChild(bar);
      });
    }
    
    // 2. Fetch popular fields ranking
    const popRep = await apiCall('/reports/popular-fields');
    const rankingList = document.getElementById('report-popular-fields-list');
    rankingList.innerHTML = '';
    
    if (popRep.length === 0) {
      rankingList.innerHTML = `<li class="list-group-item bg-transparent text-muted border-0 text-center py-4">Belum ada data booking lapangan.</li>`;
    } else {
      popRep.forEach((p, idx) => {
        const li = document.createElement('li');
        li.className = 'list-group-item bg-transparent text-white border-bottom border-secondary border-opacity-10 d-flex justify-content-between align-items-center py-3';
        
        let medal = `<span class="badge bg-secondary me-3" style="width:24px;">${idx + 1}</span>`;
        if (idx === 0) medal = `<span class="badge bg-warning text-dark me-3" style="width:24px;"><i class="fa-solid fa-trophy"></i></span>`;
        
        li.innerHTML = `
          <div class="d-flex align-items-center">
            ${medal}
            <span class="fw-semibold text-truncate" style="max-width:180px;">${p.field_name}</span>
          </div>
          <span class="badge bg-primary rounded-pill px-3 py-2">${p.count} kali disewa</span>
        `;
        rankingList.appendChild(li);
      });
    }
    
  } catch (err) {}
}


// === CORE UTILITIES ===

function formatRupiah(val) {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0
  }).format(val);
}

function formatDate(dateStr) {
  // dateStr is YYYY-MM-DD
  if (!dateStr) return '';
  const parts = dateStr.split('-');
  if (parts.length !== 3) return dateStr;
  
  const months = [
    'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
    'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
  ];
  
  const day = parseInt(parts[2]);
  const monthIdx = parseInt(parts[1]) - 1;
  const year = parts[0];
  
  return `${day} ${months[monthIdx]} ${year}`;
}

function formatMonthString(yearMonthStr) {
  // yearMonthStr is YYYY-MM
  if (!yearMonthStr) return '';
  const parts = yearMonthStr.split('-');
  if (parts.length !== 2) return yearMonthStr;
  
  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des'];
  const monthIdx = parseInt(parts[1]) - 1;
  return `${months[monthIdx]} ${parts[0]}`;
}

function calculateHourDiff(startStr, endStr) {
  try {
    const sParts = startStr.split(':');
    const eParts = endStr.split(':');
    
    const startMins = parseInt(sParts[0]) * 60 + parseInt(sParts[1]);
    const endMins = parseInt(eParts[0]) * 60 + parseInt(eParts[1]);
    
    let diffMins = endMins - startMins;
    if (diffMins < 0) diffMins += 24 * 60; // crossover midnight
    
    return (diffMins / 60).toFixed(0);
  } catch (e) {
    return '0';
  }
}
