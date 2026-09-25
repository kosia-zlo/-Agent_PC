let currentTab = 'agents';
let activeAgentId = '';

function showTab(tabId) {
    currentTab = tabId;
    document.querySelectorAll('.tab-content').forEach(el => el.classList.remove('active'));
    document.querySelectorAll('.nav-btn').forEach(el => el.classList.remove('active'));
    
    document.getElementById('tab-' + tabId).classList.add('active');
    document.getElementById('btn-' + tabId).classList.add('active');
    
    if (tabId === 'agents') loadAgents();
    if (tabId === 'tasks') loadTasks();
    if (tabId === 'audit') loadAudit();
}

async function loadAgents() {
    const res = await fetch('/api/agents');
    const agents = await res.json();
    const list = document.getElementById('agents-list');
    list.innerHTML = '';
    
    agents.forEach(a => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${a.hostname}</td>
            <td>${a.os}/${a.arch}</td>
            <td>${a.version}</td>
            <td><span class="status-pill status-${a.status}">${a.status}</span></td>
            <td>${new Date(a.last_seen).toLocaleString()}</td>
            <td><button onclick="goToTask('${a.id}')" class="btn-primary" style="font-size:0.8rem">Задача</button></td>
        `;
        list.appendChild(row);
    });
}

function goToTask(agentId) {
    activeAgentId = agentId;
    document.getElementById('task-agent-id').value = agentId;
    showTab('tasks');
}

async function loadTasks() {
    const agentId = document.getElementById('task-agent-id').value;
    const url = agentId ? `/api/tasks?agent_id=${agentId}` : '/api/tasks';
    const res = await fetch(url);
    const tasks = await res.json();
    const list = document.getElementById('tasks-list');
    list.innerHTML = '';
    
    tasks.forEach(t => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${t.id.substring(0, 8)}</td>
            <td>${t.agent_id.substring(0, 8)}</td>
            <td>${t.type}</td>
            <td><span class="status-pill status-${t.status}">${t.status}</span></td>
            <td>${new Date(t.created_at).toLocaleString()}</td>
            <td style="max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">${t.result || '-'}</td>
        `;
        list.appendChild(row);
    });
}

async function createTask() {
    const body = {
        agent_id: document.getElementById('task-agent-id').value,
        type: document.getElementById('task-type').value,
        params: document.getElementById('task-params').value,
        actor: document.getElementById('task-actor').value,
    };
    
    const res = await fetch('/api/tasks', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(body)
    });
    
    if (res.ok) {
        loadTasks();
    } else {
        const err = await res.json();
        alert('Error: ' + err.error);
    }
}

async function loadAudit() {
    const res = await fetch('/api/audit');
    const entries = await res.json();
    const list = document.getElementById('audit-list');
    list.innerHTML = '';
    
    entries.forEach(e => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${new Date(e.timestamp).toLocaleString()}</td>
            <td>${e.actor}</td>
            <td>${e.action}</td>
            <td>${e.target}</td>
            <td>${e.details}</td>
        `;
        list.appendChild(row);
    });
}

// Auto-update logic
setInterval(() => {
    if (currentTab === 'agents') loadAgents();
}, 5000);

// Init
showTab('agents');
