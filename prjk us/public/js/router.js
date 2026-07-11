// FutsalBook SPA Client-Side Router

const routes = [
  { path: '/login', view: 'view-login', auth: false },
  { path: '/register', view: 'view-register', auth: false },
  { path: '/forgot-password', view: 'view-forgot-password', auth: false },
  { path: '/reset-password', view: 'view-reset-password', auth: false },
  { path: '/dashboard', view: 'view-user-dashboard', auth: true, role: 'user' },
  { path: '/fields', view: 'view-user-fields', auth: true },
  { path: '/fields/:id', view: 'view-user-field-detail', auth: true },
  { path: '/booking/:field_id', view: 'view-user-booking-form', auth: true, role: 'user' },
  { path: '/bookings', view: 'view-user-bookings', auth: true, role: 'user' },
  { path: '/profile', view: 'view-user-profile', auth: true },
  { path: '/admin/dashboard', view: 'view-admin-dashboard', auth: true, role: 'admin' },
  { path: '/admin/fields', view: 'view-admin-fields', auth: true, role: 'admin' },
  { path: '/admin/fields/new', view: 'view-admin-field-new', auth: true, role: 'admin' },
  { path: '/admin/fields/edit/:id', view: 'view-admin-field-edit', auth: true, role: 'admin' },
  { path: '/admin/bookings', view: 'view-admin-bookings', auth: true, role: 'admin' },
  { path: '/admin/users', view: 'view-admin-users', auth: true, role: 'admin' },
  { path: '/admin/reports', view: 'view-admin-reports', auth: true, role: 'admin' }
];

// Global routing state
let currentParams = {};
let currentRoute = {};

// Match hash URL to route pattern
function matchRoute(hash) {
  if (!hash || hash === '' || hash === '#') {
    // If logged in, go to respective dashboard, otherwise login
    const user = getLoggedInUser();
    if (user) {
      return {
        route: routes.find(r => r.path === (user.role === 'admin' ? '/admin/dashboard' : '/dashboard')),
        params: {}
      };
    } else {
      return { route: routes.find(r => r.path === '/login'), params: {} };
    }
  }

  const path = hash.substring(1); // remove the '#' character

  for (const r of routes) {
    const routeParts = r.path.split('/');
    const pathParts = path.split('/');
    if (routeParts.length !== pathParts.length) continue;

    let match = true;
    let params = {};
    for (let i = 0; i < routeParts.length; i++) {
      if (routeParts[i].startsWith(':')) {
        const paramName = routeParts[i].substring(1);
        params[paramName] = pathParts[i];
      } else if (routeParts[i] !== pathParts[i]) {
        match = false;
        break;
      }
    }
    if (match) {
      return { route: r, params };
    }
  }
  return null;
}

// Check local storage for session
function getLoggedInUser() {
  try {
    const userStr = localStorage.getItem('user');
    if (!userStr) return null;
    return JSON.parse(userStr);
  } catch (e) {
    return null;
  }
}

function getAuthToken() {
  return localStorage.getItem('token');
}

// Main routing engine
function resolveRoute() {
  const hash = window.location.hash;
  const match = matchRoute(hash);

  if (!match) {
    // Route not found, redirect to home/dashboard
    window.location.hash = '#/';
    return;
  }

  const { route, params } = match;
  const user = getLoggedInUser();
  const token = getAuthToken();

  // Authentication guards
  if (route.auth && (!token || !user)) {
    // Session missing, redirect to login
    localStorage.clear();
    window.location.hash = '#/login';
    return;
  }

  if (!route.auth && token && user) {
    // Logged-in user tries to visit login/register/etc.
    window.location.hash = user.role === 'admin' ? '#/admin/dashboard' : '#/dashboard';
    return;
  }

  if (route.role && user && user.role !== route.role) {
    // Role mismatch, send back to their respective homepage
    window.location.hash = user.role === 'admin' ? '#/admin/dashboard' : '#/dashboard';
    return;
  }

  currentRoute = route;
  currentParams = params;

  // Render Layout Wrapper
  const authContainer = document.getElementById('authContainer');
  const appLayout = document.getElementById('appLayout');

  if (route.auth) {
    authContainer.classList.add('d-none');
    appLayout.classList.remove('d-none');
    
    // Update sidebar profiles
    document.getElementById('sidebar-username').innerText = user.nama;
    document.getElementById('sidebar-role').innerText = user.role;
    document.getElementById('sidebar-avatar').innerText = user.nama.charAt(0).toUpperCase();
    
    renderSidebarNavigation(user.role, route.path);
  } else {
    authContainer.classList.remove('d-none');
    appLayout.classList.add('d-none');
  }

  // Display only the active view section
  document.querySelectorAll('.view-section').forEach(view => {
    view.style.display = 'none';
  });

  const activeView = document.getElementById(route.view);
  if (activeView) {
    activeView.style.display = 'block';
  }

  // Close mobile sidebar automatically on navigation
  const sidebar = document.getElementById('appSidebar');
  if (sidebar) sidebar.classList.remove('show');

  // Trigger page-specific initializers in app.js
  triggerPageInit(route.path, params);
}

// Dynamic Sidebar navigation links based on Role
function renderSidebarNavigation(role, currentPath) {
  const navContainer = document.getElementById('sidebar-nav');
  navContainer.innerHTML = '';

  let menuItems = [];
  if (role === 'admin') {
    menuItems = [
      { name: 'Dashboard', path: '#/admin/dashboard', icon: 'fa-gauge' },
      { name: 'Kelola Lapangan', path: '#/admin/fields', icon: 'fa-layer-group' },
      { name: 'Kelola Booking', path: '#/admin/bookings', icon: 'fa-clipboard-check' },
      { name: 'Kelola User', path: '#/admin/users', icon: 'fa-users' },
      { name: 'Laporan', path: '#/admin/reports', icon: 'fa-chart-line' },
      { name: 'Profil Saya', path: '#/profile', icon: 'fa-user-gear' }
    ];
  } else {
    menuItems = [
      { name: 'Dashboard', path: '#/dashboard', icon: 'fa-gauge' },
      { name: 'Cari Lapangan', path: '#/fields', icon: 'fa-search' },
      { name: 'Riwayat Booking', path: '#/bookings', icon: 'fa-clock-rotate-left' },
      { name: 'Profil Saya', path: '#/profile', icon: 'fa-user' }
    ];
  }

  menuItems.forEach(item => {
    const link = document.createElement('a');
    link.href = item.path;
    link.className = 'nav-link-custom';
    
    // Check active
    const cleanPath = item.path.substring(1);
    const cleanCurrent = currentPath;
    if (cleanCurrent.startsWith(cleanPath) || (cleanPath === '/admin/fields' && cleanCurrent.startsWith('/admin/fields'))) {
      link.classList.add('active');
    }
    
    link.innerHTML = `<i class="fa-solid ${item.icon}"></i> <span>${item.name}</span>`;
    navContainer.appendChild(link);
  });
}

// Event Listeners for Routing
window.addEventListener('hashchange', resolveRoute);
window.addEventListener('load', resolveRoute);

function toggleMobileSidebar() {
  const sidebar = document.getElementById('appSidebar');
  sidebar.classList.toggle('show');
}
