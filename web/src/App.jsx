import { useState, useEffect, useRef } from 'react';
import { Upload, Calendar, ChevronLeft, ChevronRight, ChevronDown, Download } from 'lucide-react';
import './App.css';

// API functions
const API = {
    async getBanks() {
        const response = await fetch('/banks');
        if (!response.ok) throw new Error('Failed to fetch banks');
        return response.json();
    },

    async getUsers() {
        const response = await fetch('/users');
        if (!response.ok) throw new Error('Failed to fetch users');
        return response.json();
    },

    async uploadCSV(bankId, userId, file) {
        // Читаем файл как base64
        const reader = new FileReader();
        const base64Data = await new Promise((resolve, reject) => {
            reader.onload = (e) => {
                // Убираем префикс "data:text/csv;base64," или подобный
                const base64 = e.target.result.split(',')[1];
                resolve(base64);
            };
            reader.onerror = reject;
            reader.readAsDataURL(file);
        });

        // Создаем объект согласно protobuf структуре
        const requestData = {
            csv_data: base64Data, // Отправляем как base64 строку
            filename: file.name,
            bank_id: parseInt(bankId, 10),
            user_id: parseInt(userId, 10)
        };

        const response = await fetch('/upload-csv', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(requestData),
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Failed to upload CSV');
        }
        return response.json();
    },

    async getTransactions(month, year) {
        const response = await fetch(`/transactions?month=${month}&year=${year}`);
        if (!response.ok) throw new Error('Failed to fetch transactions');
        return response.json();
    },

    async updateTransaction(transactionId, updates) {
        const response = await fetch(`/transactions/${transactionId}`, {
            method: 'PATCH',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(updates),
        });

        if (!response.ok) throw new Error('Failed to update transaction');
        return response.json();
    },

    async getCategories() {
        const response = await fetch('/categories');
        if (!response.ok) throw new Error('Failed to fetch categories');
        return response.json();
    },

    async getTransactionTypes() {
        const response = await fetch('/transaction-types');
        if (!response.ok) throw new Error('Failed to fetch transaction types');
        return response.json();
    },

    async loadMonzoTransactions(month, year) {
        try {
            // Вычисляем даты начала и конца месяца
            const startDate = new Date(year, month - 1, 1); // month - 1, так как в JS месяцы от 0
            const endDate = new Date(year, month, 0, 23, 59, 59); // последний день месяца

            // Форматируем в ISO строки (RFC3339)
            const since = startDate.toISOString();
            const before = endDate.toISOString();

            const url = `/monzo/transactions?since=${encodeURIComponent(since)}&before=${encodeURIComponent(before)}`;
            console.log('Loading Monzo transactions for period:', { since, before });

            const response = await fetch(url);
            const data = await response.json();

            console.log('Monzo transactions response:', data);

            // Проверяем различные индикаторы того, что нужна авторизация
            if (!response.ok ||
                data.host === "api.monzo.com" ||
                !data.transactions ||
                response.status === 401) {
                return { needsAuth: true };
            }

            return data;
        } catch (error) {
            console.error('Error loading Monzo transactions:', error);
            return { needsAuth: true };
        }
    },

    async getMonzoAuthURL() {
        try {
            const response = await fetch('/monzo/auth-url');
            if (!response.ok) throw new Error('Failed to get Monzo auth URL');

            const data = await response.json();
            console.log('Auth URL response:', data);

            return data;
        } catch (error) {
            console.error('Error getting Monzo auth URL:', error);
            throw error;
        }
    },
};

// Month navigation component
function MonthNavigation({ currentDate, onDateChange }) {
    const months = [
        'January', 'February', 'March', 'April', 'May', 'June',
        'July', 'August', 'September', 'October', 'November', 'December'
    ];

    const handlePrevMonth = () => {
        const newDate = new Date(currentDate);
        newDate.setMonth(newDate.getMonth() - 1);
        onDateChange(newDate);
    };

    const handleNextMonth = () => {
        const newDate = new Date(currentDate);
        newDate.setMonth(newDate.getMonth() + 1);
        onDateChange(newDate);
    };

    const isCurrentMonth = () => {
        const now = new Date();
        return currentDate.getMonth() === now.getMonth() &&
            currentDate.getFullYear() === now.getFullYear();
    };

    return (
        <div className="month-nav">
            <button onClick={handlePrevMonth} className="nav-button">
                <ChevronLeft size={20} />
                Previous
            </button>
            <div className="month-display">
                {months[currentDate.getMonth()]} {currentDate.getFullYear()}
            </div>
            <button
                onClick={handleNextMonth}
                disabled={isCurrentMonth()}
                className="nav-button"
            >
                Next
                <ChevronRight size={20} />
            </button>
        </div>
    );
}

// Category dropdown component
function CategoryDropdown({ transaction, categories, onUpdate }) {
    const [isOpen, setIsOpen] = useState(false);
    const [isUpdating, setIsUpdating] = useState(false);
    const dropdownRef = useRef(null);

    useEffect(() => {
        const handleClickOutside = (event) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
                setIsOpen(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const handleCategorySelect = async (categoryId) => {
        if (categoryId === transaction.categoryId) {
            setIsOpen(false);
            return;
        }

        setIsUpdating(true);
        try {
            await onUpdate(transaction.id, { category_id: categoryId });
            setIsOpen(false);
        } catch (error) {
            console.error('Failed to update category:', error);
        } finally {
            setIsUpdating(false);
        }
    };

    // Use categoryName from transaction or find category by ID
    const categoryName = transaction.categoryName ||
        (transaction.categoryId && categories.find(c => c.id === transaction.categoryId)?.name) ||
        'Uncategorized';
    const isUncategorized = categoryName === 'Uncategorized';

    return (
        <div className="category-wrapper" ref={dropdownRef}>
            <button
                className={`category-button ${isUncategorized ? 'uncategorized' : ''}`}
                onClick={() => setIsOpen(!isOpen)}
                disabled={isUpdating}
            >
                {isUpdating ? 'Updating...' : categoryName}
                <ChevronDown size={16} />
            </button>
            {isOpen && (
                <div className="category-dropdown">
                    {categories.map(category => (
                        <div
                            key={category.id}
                            className={`category-option ${category.id === transaction.categoryId ? 'selected' : ''}`}
                            onClick={() => handleCategorySelect(category.id)}
                        >
                            {category.name}
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}

// Transaction type dropdown component
function TransactionTypeDropdown({ transaction, onUpdate }) {
    const [isOpen, setIsOpen] = useState(false);
    const [isUpdating, setIsUpdating] = useState(false);
    const dropdownRef = useRef(null);

    const types = [
        { value: 'INCOME', label: 'Income' },
        { value: 'OUTCOME', label: 'Outcome' }
    ];

    useEffect(() => {
        const handleClickOutside = (event) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
                setIsOpen(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const handleTypeSelect = async (type) => {
        if (type === transaction.type) {
            setIsOpen(false);
            return;
        }

        setIsUpdating(true);
        try {
            await onUpdate(transaction.id, { type });
            setIsOpen(false);
        } catch (error) {
            console.error('Failed to update type:', error);
        } finally {
            setIsUpdating(false);
        }
    };

    const currentType = types.find(t => t.value === transaction.type);

    return (
        <div className="type-wrapper" ref={dropdownRef}>
            <button
                className={`type-button ${transaction.type?.toLowerCase()}`}
                onClick={() => setIsOpen(!isOpen)}
                disabled={isUpdating}
            >
                {isUpdating ? 'Updating...' : (currentType?.label || 'Unknown')}
                <ChevronDown size={16} />
            </button>
            {isOpen && (
                <div className="type-dropdown">
                    {types.map(type => (
                        <div
                            key={type.value}
                            className={`type-option ${type.value === transaction.type ? 'selected' : ''}`}
                            onClick={() => handleTypeSelect(type.value)}
                        >
                            {type.label}
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}

// Transactions table component
function TransactionsTable({ transactions, categories, onTransactionUpdate }) {
    const [expandedCategories, setExpandedCategories] = useState(new Set());
    const [viewMode, setViewMode] = useState('categories'); // 'categories' or 'table'

    if (transactions.length === 0) {
        return (
            <div className="empty-state">
                <Calendar size={48} />
                <h3>No transactions found</h3>
                <p>Upload a CSV file or import from Monzo to see transactions for this month</p>
            </div>
        );
    }

    // Считаем общие суммы
    const totals = transactions.reduce((acc, transaction) => {
        const amount = Math.abs(parseFloat(transaction.amount));
        if (transaction.type === 'INCOME') {
            acc.income += amount;
        } else if (transaction.type === 'OUTCOME') {
            acc.outcome += amount;
        }
        return acc;
    }, { income: 0, outcome: 0 });

    // Группируем транзакции по категориям
    const transactionsByCategory = transactions.reduce((acc, transaction) => {
        // Используем categoryId или 'uncategorized' если его нет
        const categoryId = transaction.categoryId ? String(transaction.categoryId) : 'uncategorized';

        if (!acc[categoryId]) {
            acc[categoryId] = {
                transactions: [],
                income: 0,
                outcome: 0
            };
        }

        acc[categoryId].transactions.push(transaction);

        // Считаем суммы только для известных типов
        const amount = Math.abs(parseFloat(transaction.amount));
        if (transaction.type === 'INCOME') {
            acc[categoryId].income += amount;
        } else if (transaction.type === 'OUTCOME') {
            acc[categoryId].outcome += amount;
        }

        return acc;
    }, {});

    const toggleCategory = (categoryId) => {
        const newExpanded = new Set(expandedCategories);
        if (newExpanded.has(categoryId)) {
            newExpanded.delete(categoryId);
        } else {
            newExpanded.add(categoryId);
        }
        setExpandedCategories(newExpanded);
    };

    const formatDate = (dateString) => {
        if (!dateString) return 'Invalid Date';
        const date = new Date(dateString);
        if (isNaN(date.getTime())) return 'Invalid Date';
        const day = date.getDate();
        const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
        const month = months[date.getMonth()];
        return `${day}\u00A0${month}`; // Используем неразрывный пробел
    };

    const formatAmount = (amount, type) => {
        const value = Math.abs(parseFloat(amount)); // Используем Math.abs чтобы убрать возможный минус из значения
        const formatted = value.toFixed(2);

        if (type === 'INCOME') {
            return `+${formatted}`;
        } else if (type === 'OUTCOME') {
            return `-${formatted}`;
        } else {
            // Для unknown типа - без знака
            return formatted;
        }
    };

    const getCategoryName = (categoryId) => {
        if (categoryId === 'uncategorized') return 'Uncategorized';

        // Пробуем найти категорию, преобразуя categoryId в число
        const category = categories.find(c => c.id === parseInt(categoryId) || c.id === categoryId);

        if (category) {
            return category.name;
        }

        // Если не нашли, возвращаем дефолтное имя
        return `Category ${categoryId}`;
    };

    return (
        <div>
            {/* Общая статистика за месяц */}
            <div className="month-summary">
                <div className="summary-card">
                    <h4>Total Income</h4>
                    <span className="summary-income">+{totals.income.toFixed(2)}</span>
                </div>
                <div className="summary-card">
                    <h4>Total Expenses</h4>
                    <span className="summary-outcome">-{totals.outcome.toFixed(2)}</span>
                </div>
                <div className="summary-card">
                    <h4>Balance</h4>
                    <span className={totals.income - totals.outcome >= 0 ? 'summary-income' : 'summary-outcome'}>
                        {(totals.income - totals.outcome).toFixed(2)}
                    </span>
                </div>
            </div>

            {/* Вкладки */}
            <div className="view-tabs">
                <button
                    className={`tab-button ${viewMode === 'categories' ? 'active' : ''}`}
                    onClick={() => setViewMode('categories')}
                >
                    By Categories
                </button>
                <button
                    className={`tab-button ${viewMode === 'table' ? 'active' : ''}`}
                    onClick={() => setViewMode('table')}
                >
                    Table View
                </button>
            </div>

            {/* Отображение по категориям */}
            {viewMode === 'categories' && (
                <div className="categories-container">
                    {Object.entries(transactionsByCategory)
                        .sort(([, a], [, b]) => b.outcome - a.outcome) // Сортируем по убыванию расходов
                        .map(([categoryId, data]) => (
                            <div key={categoryId} className="category-section">
                                <div
                                    className="category-header"
                                    onClick={() => toggleCategory(categoryId)}
                                >
                                    <div className="category-title">
                                        <ChevronRight
                                            size={20}
                                            className={`chevron ${expandedCategories.has(categoryId) ? 'expanded' : ''}`}
                                        />
                                        <h3>{getCategoryName(categoryId)}</h3>
                                        <span className="transaction-count">({data.transactions.length})</span>
                                    </div>
                                    <div className="category-summary">
                                        {data.income > 0 && (
                                            <span className="summary-income">Income: +{data.income.toFixed(2)}</span>
                                        )}
                                        {data.outcome > 0 && (
                                            <span className="summary-outcome">Expenses: -{data.outcome.toFixed(2)}</span>
                                        )}
                                    </div>
                                </div>

                                {expandedCategories.has(categoryId) && (
                                    <div className="category-transactions">
                                        <table>
                                            <thead>
                                            <tr>
                                                <th>Date</th>
                                                <th>Description</th>
                                                <th>Amount</th>
                                                <th>Type</th>
                                                <th>Category</th>
                                                <th>Bank</th>
                                                <th>User</th>
                                            </tr>
                                            </thead>
                                            <tbody>
                                            {data.transactions
                                                .sort((a, b) => new Date(b.transactionDate) - new Date(a.transactionDate))
                                                .map(transaction => (
                                                    <tr key={transaction.id}>
                                                        <td>{formatDate(transaction.transactionDate)}</td>
                                                        <td>{transaction.description}</td>
                                                        <td className={`amount ${transaction.type?.toLowerCase()}`}>
                                                            {formatAmount(transaction.amount, transaction.type)}
                                                        </td>
                                                        <td>
                                                            <TransactionTypeDropdown
                                                                transaction={transaction}
                                                                onUpdate={onTransactionUpdate}
                                                            />
                                                        </td>
                                                        <td>
                                                            <CategoryDropdown
                                                                transaction={transaction}
                                                                categories={categories}
                                                                onUpdate={onTransactionUpdate}
                                                            />
                                                        </td>
                                                        <td>{transaction.bankName || 'Unknown'}</td>
                                                        <td>{transaction.userName || '-'}</td>
                                                    </tr>
                                                ))}
                                            </tbody>
                                        </table>
                                    </div>
                                )}
                            </div>
                        ))}
                </div>
            )}

            {/* Табличное отображение */}
            {viewMode === 'table' && (
                <div className="table-wrapper">
                    <table>
                        <thead>
                        <tr>
                            <th>User</th>
                            <th>Bank</th>
                            <th>Date</th>
                            <th>Amount</th>
                            <th>Type</th>
                            <th>Category</th>
                            <th>Description</th>
                            <th>External ID</th>
                        </tr>
                        </thead>
                        <tbody>
                        {transactions.map(transaction => (
                            <tr key={transaction.id}>
                                <td>{transaction.userName || '-'}</td>
                                <td>{transaction.bankName || 'Unknown'}</td>
                                <td>{formatDate(transaction.transactionDate)}</td>
                                <td className={`amount ${transaction.type?.toLowerCase()}`}>
                                    {formatAmount(transaction.amount, transaction.type)}
                                </td>
                                <td>
                                    <TransactionTypeDropdown
                                        transaction={transaction}
                                        onUpdate={onTransactionUpdate}
                                    />
                                </td>
                                <td>
                                    <CategoryDropdown
                                        transaction={transaction}
                                        categories={categories}
                                        onUpdate={onTransactionUpdate}
                                    />
                                </td>
                                <td>{transaction.description}</td>
                                <td className="external-id">{transaction.externalId || '-'}</td>
                            </tr>
                        ))}
                        </tbody>
                    </table>
                </div>
            )}
        </div>
    );
}

// Upload section component
function UploadSection({ banks, users, onUpload, onMonzoImport }) {
    const [selectedBank, setSelectedBank] = useState('');
    const [selectedUser, setSelectedUser] = useState('');
    const [file, setFile] = useState(null);
    const [isUploading, setIsUploading] = useState(false);
    const [isImportingMonzo, setIsImportingMonzo] = useState(false);

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (!selectedBank || !selectedUser || !file) return;

        setIsUploading(true);
        try {
            await onUpload(selectedBank, selectedUser, file);
            // Reset form
            setSelectedBank('');
            setSelectedUser('');
            setFile(null);
            e.target.reset();
        } catch (error) {
            console.error('Upload failed:', error);
        } finally {
            setIsUploading(false);
        }
    };

    const handleMonzoImport = async () => {
        setIsImportingMonzo(true);
        try {
            await onMonzoImport();
        } catch (error) {
            console.error('Monzo import failed:', error);
        } finally {
            setIsImportingMonzo(false);
        }
    };

    return (
        <div className="upload-section">
            <h2>
                <Upload size={24} />
                Upload Transactions
            </h2>
            <form onSubmit={handleSubmit}>
                <div className="upload-controls">
                    <div className="form-group">
                        <label htmlFor="user">User</label>
                        <select
                            id="user"
                            value={selectedUser}
                            onChange={(e) => setSelectedUser(e.target.value)}
                            required
                        >
                            <option value="">Select a user</option>
                            {Array.isArray(users) && users.map(user => (
                                <option key={user.id} value={user.id}>
                                    {user.name}
                                </option>
                            ))}
                        </select>
                    </div>
                    <div className="form-group">
                        <label htmlFor="bank">Bank</label>
                        <select
                            id="bank"
                            value={selectedBank}
                            onChange={(e) => setSelectedBank(e.target.value)}
                            required
                        >
                            <option value="">Select a bank</option>
                            {Array.isArray(banks) && banks.map(bank => (
                                <option key={bank.id} value={bank.id}>
                                    {bank.name}
                                </option>
                            ))}
                        </select>
                    </div>
                    <div className="form-group">
                        <label htmlFor="file">CSV File</label>
                        <input
                            id="file"
                            type="file"
                            accept=".csv"
                            onChange={(e) => setFile(e.target.files[0])}
                            required
                        />
                    </div>
                    <button
                        type="submit"
                        className="upload-btn"
                        disabled={isUploading || !selectedBank || !selectedUser || !file}
                    >
                        {isUploading ? 'Uploading...' : 'Upload CSV'}
                    </button>
                </div>
            </form>
            <div className="monzo-section">
                <button
                    onClick={handleMonzoImport}
                    className="monzo-btn"
                    disabled={isImportingMonzo}
                >
                    <Download size={20} />
                    {isImportingMonzo ? 'Importing from Monzo...' : 'Import from Monzo'}
                </button>
            </div>
        </div>
    );
}

// Main App component
function App() {
    const [currentDate, setCurrentDate] = useState(new Date());
    const [transactions, setTransactions] = useState([]);
    const [categories, setCategories] = useState([]);
    const [banks, setBanks] = useState([]);
    const [users, setUsers] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [success, setSuccess] = useState(null);

    // Load initial data
    useEffect(() => {
        const loadInitialData = async () => {
            try {
                const [banksData, usersData, categoriesData] = await Promise.all([
                    API.getBanks(),
                    API.getUsers(),
                    API.getCategories()
                ]);
                // Extract arrays from response objects
                setBanks(banksData.banks || []);
                setUsers(usersData.users || []);
                // Handle both possible formats for categories
                setCategories(categoriesData.categories || categoriesData.category || []);
            } catch (err) {
                setError('Failed to load initial data');
                console.error(err);
            }
        };

        loadInitialData();
    }, []);

    // Load transactions when date changes
    useEffect(() => {
        const loadTransactions = async () => {
            setLoading(true);
            setError(null);
            try {
                const data = await API.getTransactions(
                    currentDate.getMonth() + 1,
                    currentDate.getFullYear()
                );
                setTransactions(data.transactions || []);
            } catch (err) {
                setError('Failed to load transactions');
                console.error(err);
            } finally {
                setLoading(false);
            }
        };

        loadTransactions();
    }, [currentDate]);

    const handleUpload = async (bankId, userId, file) => {
        setError(null);
        setSuccess(null);
        try {
            const result = await API.uploadCSV(bankId, userId, file);
            setSuccess(`Successfully uploaded ${result.count || 0} transactions`);

            // Reload transactions for current month
            const data = await API.getTransactions(
                currentDate.getMonth() + 1,
                currentDate.getFullYear()
            );
            setTransactions(data.transactions || []);
        } catch (err) {
            setError('Failed to upload CSV file');
            throw err;
        }
    };

    const handleTransactionUpdate = async (transactionId, updates) => {
        try {
            const result = await API.updateTransaction(transactionId, updates);

            // Update local state
            setTransactions(prev => prev.map(t => {
                if (t.id === transactionId) {
                    const updated = { ...t };
                    if (updates.category_id !== undefined) {
                        updated.categoryId = updates.category_id;
                        updated.categoryName = categories.find(c => c.id === updates.category_id)?.name || 'Uncategorized';
                    }
                    if (updates.type !== undefined) {
                        updated.type = updates.type;
                    }
                    return updated;
                }
                return t;
            }));
        } catch (err) {
            setError('Failed to update transaction');
            throw err;
        }
    };

    const handleMonzoImport = async () => {
        setError(null);
        setSuccess(null);
        try {
            console.log('Starting Monzo import...');
            // Передаем текущий выбранный месяц и год
            const month = currentDate.getMonth() + 1; // +1 так как getMonth() возвращает 0-11
            const year = currentDate.getFullYear();

            const result = await API.loadMonzoTransactions(month, year);

            if (result.needsAuth) {
                console.log('Need auth, getting auth URL...');
                // Пользователь не авторизован, получаем URL для авторизации
                const authData = await API.getMonzoAuthURL();

                // Проверяем разные возможные форматы ответа
                const authUrl = authData.authUrl || authData.auth_url || authData.url;
                console.log('Auth URL to redirect:', authUrl);

                if (authUrl) {
                    // Перенаправляем на страницу авторизации Monzo
                    window.location.href = authUrl;
                } else {
                    console.error('No auth URL found in response:', authData);
                    setError('Failed to get Monzo authorization URL');
                }
                return;
            }

            // Проверяем, есть ли транзакции в ответе
            if (!result.transactions || result.transactions.length === 0) {
                setSuccess('No transactions found in Monzo');
                return;
            }

            // Успешно загрузили транзакции
            console.log('Transactions loaded successfully:', result.transactions.length);
            setSuccess(`Successfully imported ${result.transactions.length} transactions from Monzo`);

            // Reload transactions for current month
            const data = await API.getTransactions(
                currentDate.getMonth() + 1,
                currentDate.getFullYear()
            );
            setTransactions(data.transactions || []);
        } catch (err) {
            console.error('Monzo import error:', err);
            setError('Failed to import from Monzo: ' + err.message);
        }
    };

    // Check if we returned from Monzo auth
    useEffect(() => {
        const urlParams = new URLSearchParams(window.location.search);
        const code = urlParams.get('code');
        const state = urlParams.get('state');

        if (code && state) {
            // We returned from Monzo auth, clear the URL
            window.history.replaceState({}, document.title, window.location.pathname);

            // Try to load Monzo transactions again
            setSuccess('Successfully authorized with Monzo. Loading transactions...');
            handleMonzoImport();
        }
    }, [currentDate]); // Добавляем currentDate в зависимости

    return (
        <div className="container">
            <h1>Transaction Manager</h1>

            {error && (
                <div className="message error">
                    {error}
                    <button
                        onClick={() => setError(null)}
                        className="close-button"
                    >
                        ×
                    </button>
                </div>
            )}

            {success && (
                <div className="message success">
                    {success}
                    <button
                        onClick={() => setSuccess(null)}
                        className="close-button"
                    >
                        ×
                    </button>
                </div>
            )}

            <UploadSection
                banks={banks}
                users={users}
                onUpload={handleUpload}
                onMonzoImport={handleMonzoImport}
            />

            <MonthNavigation
                currentDate={currentDate}
                onDateChange={setCurrentDate}
            />

            <div className="transactions-section">
                {loading ? (
                    <div className="loading">Loading transactions...</div>
                ) : (
                    <TransactionsTable
                        transactions={transactions}
                        categories={categories}
                        onTransactionUpdate={handleTransactionUpdate}
                    />
                )}
            </div>
        </div>
    );
}

export default App;