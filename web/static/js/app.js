/**
 * 前端应用主逻辑
 * 处理UI交互和API调用
 */

const API_BASE = '/api/v1';
const API_TOKEN = prompt('请输入API Token:') || localStorage.getItem('api_token') || '';

if (API_TOKEN) {
    localStorage.setItem('api_token', API_TOKEN);
}

let currentPage = 1;
let currentLimit = 20;
let isSearchMode = false;
let searchQuery = '';

// 初始化
document.addEventListener('DOMContentLoaded', function() {
    loadLinks();
    refreshStats();
});

/**
 * 加载链接列表
 */
async function loadLinks(page = 1) {
    currentPage = page;
    const tbody = document.getElementById('linksTableBody');
    tbody.innerHTML = '<tr><td colspan="6" class="loading">加载中...</td></tr>';

    try {
        let url;
        if (isSearchMode && searchQuery) {
            url = `${API_BASE}/links/search?q=${encodeURIComponent(searchQuery)}&page=${page}&limit=${currentLimit}`;
        } else {
            url = `${API_BASE}/links?page=${page}&limit=${currentLimit}`;
        }

        const response = await fetch(url, {
            headers: {
                'Authorization': `Bearer ${API_TOKEN}`
            }
        });

        if (!response.ok) {
            throw new Error('加载失败');
        }

        const data = await response.json();
        displayLinks(data.links || []);
        displayPagination(data);
    } catch (error) {
        tbody.innerHTML = `<tr><td colspan="6" class="loading">加载失败: ${error.message}</td></tr>`;
        console.error('加载链接失败:', error);
    }
}

/**
 * 显示链接列表
 */
function displayLinks(links) {
    const tbody = document.getElementById('linksTableBody');
    
    if (links.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="loading">暂无数据</td></tr>';
        return;
    }

    tbody.innerHTML = links.map(link => `
        <tr>
            <td><a href="/${link.code}" target="_blank" class="short-url">${link.code}</a></td>
            <td class="url-cell" title="${link.original_url}">${link.original_url}</td>
            <td>${link.title || '-'}</td>
            <td>${link.click_count || 0}</td>
            <td>${formatDate(link.created_at)}</td>
            <td>
                <button class="btn btn-danger" onclick="deleteLink('${link.code}')">删除</button>
            </td>
        </tr>
    `).join('');
}

/**
 * 显示分页
 */
function displayPagination(data) {
    const pagination = document.getElementById('pagination');
    
    if (!data.total_pages || data.total_pages <= 1) {
        pagination.innerHTML = '';
        return;
    }

    let html = '';
    for (let i = 1; i <= data.total_pages; i++) {
        html += `<button class="${i === data.page ? 'active' : ''}" onclick="loadLinks(${i})">${i}</button>`;
    }
    pagination.innerHTML = html;
}

/**
 * 刷新统计信息
 */
async function refreshStats() {
    try {
        const response = await fetch(`${API_BASE}/stats`, {
            headers: {
                'Authorization': `Bearer ${API_TOKEN}`
            }
        });

        if (!response.ok) {
            throw new Error('加载统计失败');
        }

        const data = await response.json();
        document.getElementById('totalLinks').textContent = data.total_links || 0;
        document.getElementById('totalClicks').textContent = data.total_clicks || 0;
        document.getElementById('todayClicks').textContent = data.today_clicks || 0;
    } catch (error) {
        console.error('加载统计失败:', error);
    }
}

/**
 * 创建链接
 */
async function createLink(event) {
    event.preventDefault();
    
    const originalUrl = document.getElementById('originalUrl').value;
    const title = document.getElementById('linkTitle').value;
    const customCode = document.getElementById('customCode').value;

    const data = {
        url: originalUrl,
        title: title
    };

    if (customCode) {
        data.code = customCode;
    }

    try {
        const response = await fetch(`${API_BASE}/links`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${API_TOKEN}`
            },
            body: JSON.stringify(data)
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || '创建失败');
        }

        const result = await response.json();
        alert('创建成功！短链接: ' + result.short_url);
        closeCreateModal();
        loadLinks();
        refreshStats();
    } catch (error) {
        alert('创建失败: ' + error.message);
        console.error('创建链接失败:', error);
    }
}

/**
 * 删除链接
 */
async function deleteLink(code) {
    if (!confirm('确定要删除这个链接吗？')) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/links/${code}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${API_TOKEN}`
            }
        });

        if (!response.ok) {
            throw new Error('删除失败');
        }

        alert('删除成功');
        loadLinks();
        refreshStats();
    } catch (error) {
        alert('删除失败: ' + error.message);
        console.error('删除链接失败:', error);
    }
}

/**
 * 搜索处理
 */
function handleSearch(event) {
    if (event.key === 'Enter') {
        performSearch();
    }
}

function performSearch() {
    const query = document.getElementById('searchInput').value.trim();
    searchQuery = query;
    isSearchMode = query !== '';
    loadLinks(1);
}

/**
 * 显示创建模态框
 */
function showCreateModal() {
    document.getElementById('createModal').style.display = 'block';
    document.getElementById('createForm').reset();
}

/**
 * 关闭创建模态框
 */
function closeCreateModal() {
    document.getElementById('createModal').style.display = 'none';
}

/**
 * 格式化日期
 */
function formatDate(dateString) {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return date.toLocaleString('zh-CN');
}

// 点击模态框外部关闭
window.onclick = function(event) {
    const modal = document.getElementById('createModal');
    if (event.target === modal) {
        closeCreateModal();
    }
}

