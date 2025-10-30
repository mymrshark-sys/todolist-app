// Global variables

let currentFilter = 'all';
let editModal;
let currentUser = null;

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    // Initialize Bootstrap modal
    editModal = new bootstrap.Modal(document.getElementById('editModal'));

    // Load user info and notes
    getCurrentUser().then(() => {
        loadNotes();
    });

    // Setup event listeners
    setupEventListeners();

    // Setup logout button
    document.getElementById('logoutBtn').addEventListener('click', handleLogout);
});

// Get current user info
async function getCurrentUser() {
    try {
        const response = await fetch('/api/user');
        if (!response.ok) {
            throw new Error('Not authenticated');
        }
        currentUser = await response.json();
        
        // Update UI with user info
        const userDisplayName = document.getElementById('userDisplayName');
        if (userDisplayName) {
            userDisplayName.textContent = currentUser.full_name || currentUser.username;
        }
    } catch (error) {
        console.error('Error getting user:', error);
        window.location.href = '/login';
    }
}

// Setup all event listeners
function setupEventListeners() {
    // Form submit for adding new note
    document.getElementById('noteForm').addEventListener('submit', handleAddNote);
    
    // Filter tabs
    document.querySelectorAll('#filterTabs .nav-link').forEach(tab => {
        tab.addEventListener('click', function() {
            // Update active tab
            document.querySelectorAll('#filterTabs .nav-link').forEach(t => t.classList.remove('active'));
            this.classList.add('active');
            
            // Update filter and reload notes
            currentFilter = this.dataset.filter;
            loadNotes();
        });
    });
    
    // Save edit button
    document.getElementById('saveEditBtn').addEventListener('click', handleSaveEdit);
}

// Load all notes from API
async function loadNotes() {
    try {
        const response = await fetch('/api/notes');
        if (!response.ok) throw new Error('Failed to load notes');
        
        const notes = await response.json();
        displayNotes(notes);
    } catch (error) {
        console.error('Error loading notes:', error);
        showNotification('Failed to load notes', 'danger');
    }
}

// Display notes in the UI
function displayNotes(notes) {
    const notesList = document.getElementById('notesList');
    const emptyState = document.getElementById('emptyState');
    
    // Filter notes based on current filter
    let filteredNotes = notes;
    if (currentFilter !== 'all') {
        filteredNotes = notes.filter(note => note.status === currentFilter);
    }
    
    // Show empty state if no notes
    if (filteredNotes.length === 0) {
        notesList.innerHTML = '';
        emptyState.style.display = 'block';
        return;
    }
    
    emptyState.style.display = 'none';
    
    // Generate HTML for each note
    notesList.innerHTML = filteredNotes.map(note => createNoteCard(note)).join('');
    
    // Add event listeners to action buttons
    attachNoteEventListeners();
}

// Create HTML for a single note card
function createNoteCard(note) {
    const isCompleted = note.status === 'completed';
    const statusBadge = isCompleted 
        ? '<span class="badge bg-success"><i class="bi bi-check-circle"></i> Completed</span>'
        : '<span class="badge bg-warning"><i class="bi bi-clock"></i> Pending</span>';
    
    const completedClass = isCompleted ? 'note-completed' : '';
    const toggleIcon = isCompleted ? 'bi-arrow-counterclockwise' : 'bi-check-circle';
    const toggleText = isCompleted ? 'Reopen' : 'Complete';
    
    const createdDate = new Date(note.created_at).toLocaleDateString('id-ID', {
        day: 'numeric',
        month: 'long',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
    
    return `
        <div class="col-md-6 col-lg-4 note-card">
            <div class="card shadow-sm ${completedClass}">
                <div class="card-body">
                    <div class="d-flex justify-content-between align-items-start mb-2">
                        <h5 class="note-title">${escapeHtml(note.title)}</h5>
                        ${statusBadge}
                    </div>
                    <p class="note-content">${escapeHtml(note.content)}</p>
                    <div class="note-meta">
                        <i class="bi bi-calendar"></i> ${createdDate}
                    </div>
                    <div class="action-buttons">
                        <button class="btn btn-sm btn-success toggle-btn" data-id="${note.id}">
                            <i class="bi ${toggleIcon}"></i> ${toggleText}
                        </button>
                        <button class="btn btn-sm btn-warning edit-btn" data-id="${note.id}">
                            <i class="bi bi-pencil"></i> Edit
                        </button>
                        <button class="btn btn-sm btn-danger delete-btn" data-id="${note.id}">
                            <i class="bi bi-trash"></i> Delete
                        </button>
                    </div>
                </div>
            </div>
        </div>
    `;
}

// Handle adding new note
async function handleAddNote(e) {
    e.preventDefault();
    
    const title = document.getElementById('noteTitle').value.trim();
    const content = document.getElementById('noteContent').value.trim();
    
    if (!title || !content) {
        showNotification('Please fill in all fields', 'warning');
        return;
    }
    
    try {
        const response = await fetch('/api/notes', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ title, content })
        });
        
        if (!response.ok) throw new Error('Failed to create note');
        
        // Clear form
        document.getElementById('noteForm').reset();
        
        // Reload notes
        await loadNotes();
        
        showNotification('Task added successfully!', 'success');
    } catch (error) {
        console.error('Error creating note:', error);
        showNotification('Failed to add task', 'danger');
    }
}

// Toggle note status (pending/completed)
async function toggleNoteStatus(id) {
    try {
        const response = await fetch(`/api/notes/${id}/toggle`, {
            method: 'PATCH'
        });
        
        if (!response.ok) throw new Error('Failed to toggle status');
        
        await loadNotes();
        showNotification('Status updated!', 'success');
    } catch (error) {
        console.error('Error toggling status:', error);
        showNotification('Failed to update status', 'danger');
    }
}

// Open edit modal with note data
async function openEditModal(id) {
    try {
        const response = await fetch('/api/notes');
        if (!response.ok) throw new Error('Failed to load note');
        
        const notes = await response.json();
        const note = notes.find(n => n.id === parseInt(id));
        
        if (!note) throw new Error('Note not found');
        
        // Fill modal form
        document.getElementById('editNoteId').value = note.id;
        document.getElementById('editNoteTitle').value = note.title;
        document.getElementById('editNoteContent').value = note.content;
        
        // Show modal
        editModal.show();
    } catch (error) {
        console.error('Error loading note:', error);
        showNotification('Failed to load task', 'danger');
    }
}

// Handle saving edited note
async function handleSaveEdit() {
    const id = document.getElementById('editNoteId').value;
    const title = document.getElementById('editNoteTitle').value.trim();
    const content = document.getElementById('editNoteContent').value.trim();
    
    if (!title || !content) {
        showNotification('Please fill in all fields', 'warning');
        return;
    }
    
    try {
        const response = await fetch(`/api/notes/${id}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ title, content })
        });
        
        if (!response.ok) throw new Error('Failed to update note');
        
        // Close modal
        editModal.hide();
        
        // Reload notes
        await loadNotes();
        
        showNotification('Task updated successfully!', 'success');
    } catch (error) {
        console.error('Error updating note:', error);
        showNotification('Failed to update task', 'danger');
    }
}

// Delete note
async function deleteNote(id) {
    try {
        const response = await fetch(`/api/notes/${id}`, {
            method: 'DELETE'
        });
        
        if (!response.ok) throw new Error('Failed to delete note');
        
        await loadNotes();
        showNotification('Task deleted successfully!', 'success');
    } catch (error) {
        console.error('Error deleting note:', error);
        showNotification('Failed to delete task', 'danger');
    }
}

// Show notification toast
function showNotification(message, type = 'info') {
    // Create toast element
    const toast = document.createElement('div');
    toast.className = `alert alert-${type} alert-dismissible fade show position-fixed top-0 start-50 translate-middle-x mt-3`;
    toast.style.zIndex = '9999';
    toast.innerHTML = `
        ${message}
        <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
    `;
    
    document.body.appendChild(toast);
    
    // Auto remove after 3 seconds
    setTimeout(() => {
        toast.remove();
    }, 3000);
}

// Escape HTML to prevent XSS
function escapeHtml(text) {
    const map = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#039;'
    };
    return text ? text.replace(/[&<>"']/g, m => map[m]) : '';
}

// Handle logout
async function handleLogout() {
    try {
        const response = await fetch('/api/logout', {
            method: 'POST'
        });
        
        if (response.ok) {
            window.location.href = '/login';
        }
    } catch (error) {
        console.error('Error logging out:', error);
    }
}

// Attach event listeners to note buttons
function attachNoteEventListeners() {
    // Toggle status buttons
    document.querySelectorAll('.toggle-btn').forEach(btn => {
        btn.addEventListener('click', () => toggleNoteStatus(btn.dataset.id));
    });

    // Edit buttons
    document.querySelectorAll('.edit-btn').forEach(btn => {
        btn.addEventListener('click', () => openEditModal(btn.dataset.id));
    });

    // Delete buttons
    document.querySelectorAll('.delete-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            if (confirm('Are you sure you want to delete this task?')) {
                deleteNote(btn.dataset.id);
            }
        });
    });
}