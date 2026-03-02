const output = document.getElementById("output");
const tokenState = document.getElementById("tokenState");
const themeBtn = document.getElementById("themeBtn");
let token = "";

const byId = (id) => document.getElementById(id);

function print(data) {
  output.textContent = JSON.stringify(data, null, 2);
}

function setTheme(theme) {
  document.body.classList.toggle("dark", theme === "dark");
  if (themeBtn) {
    themeBtn.textContent = theme === "dark" ? "☀️ Light" : "🌙 Dark";
  }
  localStorage.setItem("ui-theme", theme);
}

function initTheme() {
  const saved = localStorage.getItem("ui-theme");
  setTheme(saved === "dark" ? "dark" : "light");
}

function authHeader() {
  return token ? { Authorization: `Bearer ${token}` } : {};
}

async function api(path, options = {}, useAuth = true) {
  const response = await fetch(path, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
      ...(useAuth ? authHeader() : {}),
    },
  });

  const body = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw { status: response.status, body };
  }
  return body;
}

async function signIn() {
  try {
    const result = await api("/auth/sign-in", {
      method: "POST",
      body: JSON.stringify({
        email: byId("email").value,
        password: byId("password").value,
      }),
      headers: {},
    }, false);

    token = result.token || "";
    tokenState.textContent = token ? "Authenticated" : "No token";
    print(result);
  } catch (err) {
    print(err);
  }
}

function userRow(user) {
  const tr = document.createElement("tr");
  tr.innerHTML = `
    <td>${user.id}</td>
    <td>${user.email}</td>
    <td>${user.name}</td>
    <td class="actions">
      <button data-id="${user.id}" data-action="load">Load</button>
      <button data-id="${user.id}" data-action="get">Get</button>
    </td>
  `;
  return tr;
}

async function listUsers() {
  try {
    const q = encodeURIComponent(byId("searchEmail").value.trim());
    const result = await api(`/users${q ? `?email=${q}` : ""}`);

    const tbody = byId("usersBody");
    tbody.innerHTML = "";
    (result.users || []).forEach((user) => tbody.appendChild(userRow(user)));
    print(result);
  } catch (err) {
    print(err);
  }
}

async function getUser() {
  const id = byId("userId").value;
  if (!id) {
    print({ error: "user id is required" });
    return;
  }

  try {
    const result = await api(`/users/${id}`);
    byId("userName").value = result.name || "";
    print(result);
  } catch (err) {
    print(err);
  }
}

async function updateUser() {
  const id = byId("userId").value;
  const name = byId("userName").value;

  if (!id) {
    print({ error: "user id is required" });
    return;
  }

  try {
    const result = await api(`/users/${id}`, {
      method: "PUT",
      body: JSON.stringify({ name }),
    });
    print(result);
    await listUsers();
  } catch (err) {
    print(err);
  }
}

byId("signInBtn").addEventListener("click", signIn);
byId("listBtn").addEventListener("click", listUsers);
byId("getBtn").addEventListener("click", getUser);
byId("updateBtn").addEventListener("click", updateUser);

if (themeBtn) {
  themeBtn.addEventListener("click", () => {
    const isDark = document.body.classList.contains("dark");
    setTheme(isDark ? "light" : "dark");
  });
}

byId("usersBody").addEventListener("click", async (event) => {
  const target = event.target;
  if (!(target instanceof HTMLButtonElement)) return;

  const id = target.getAttribute("data-id");
  const action = target.getAttribute("data-action");
  if (!id || !action) return;

  byId("userId").value = id;

  if (action === "load") {
    try {
      const result = await api(`/users/${id}`);
      byId("userName").value = result.name || "";
      print(result);
    } catch (err) {
      print(err);
    }
    return;
  }

  if (action === "get") {
    await getUser();
  }
});

print({ message: "Sign in to start." });
initTheme();
