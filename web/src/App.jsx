import { useState, useEffect } from 'react';
import UploadTransactions from './UploadTransactions.jsx';
import TransactionsList from './TransactionsList.jsx';
import ChartsView from './ChartsView.jsx';
import './App.css';

// Monzo Auth Modal Component
function MonzoAuthModal({ onConfirm, onClose, error }) {
    const [isLoading, setIsLoading] = useState(false);
    const [modalError, setModalError] = useState(error);

    const handleConfirm = async () => {
        setIsLoading(true);
        setModalError(null);
        try {
            const success = await onConfirm();
            if (!success) {
                setModalError('Failed to confirm Monzo account access');
            }
        } catch (err) {
            setModalError(err.message || 'Failed to confirm Monzo account');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="modal-overlay">
            <div className="modal-content">
                <button className="modal-close" onClick={onClose}>×</button>
                <h3>Allow access to your Monzo data!</h3>
                {modalError && (
                    <div className="modal-error">
                        {modalError}
                    </div>
                )}
                <div className="modal-actions">
                    <button
                        className="modal-button confirm"
                        onClick={handleConfirm}
                        disabled={isLoading}
                    >
                        {isLoading ? 'Confirming...' : 'OK'}
                    </button>
                </div>
            </div>
        </div>
    );
}

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
        const reader = new FileReader();
        const base64Data = await new Promise((resolve, reject) => {
            reader.onload = (e) => {
                const base64 = e.target.result.split(',')[1];
                resolve(base64);
            };
            reader.onerror = reject;
            reader.readAsDataURL(file);
        });

        const requestData = {
            csv_data: base64Data,
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

    async loadMonzoTransactions(month, year, userId, bankId) {
        try {
            const startDate = new Date(year, month - 1, 1);
            const endDate = new Date(year, month, 0, 23, 59, 59);

            const since = startDate.toISOString();
            const before = endDate.toISOString();

            const url = `/monzo/transactions?since=${encodeURIComponent(since)}&before=${encodeURIComponent(before)}&user_id=${userId}&bank_id=${bankId}`;
            console.log('Loading Monzo transactions with params:', { since, before, userId, bankId });

            const response = await fetch(url);

            if (!response.ok) {
                const errorText = await response.text();
                console.error('Monzo API error response:', errorText);

                // Check if error is Unauthenticated
                if (response.status === 401 || errorText.includes('Unauthenticated')) {
                    return { needsAuth: true };
                }

                throw new Error('Failed to load transactions: ' + errorText);
            }

            const data = await response.json();
            console.log('Monzo transactions response:', data);

            // New response format only has success boolean
            return data;
        } catch (error) {
            console.error('Error loading Monzo transactions:', error);
            throw error;
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

    async callMonzoCallback(code, state) {
        try {
            const url = `/monzo/callback?code=${encodeURIComponent(code)}&state=${encodeURIComponent(state)}`;

            console.log('Calling Monzo callback URL:', url);

            const response = await fetch(url, {
                method: 'GET',
                headers: {
                    'Accept': 'application/json',
                }
            });

            console.log('Response status:', response.status);

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            console.log('Monzo callback response:', data);

            return data;
        } catch (error) {
            console.error('Error calling Monzo callback:', error);
            throw error;
        }
    },

    async confirmMonzoAccount() {
        try {
            const response = await fetch('/monzo/account');
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(errorText || 'Failed to confirm Monzo account');
            }

            const data = await response.json();
            return data;
        } catch (error) {
            console.error('Error confirming Monzo account:', error);
            throw error;
        }
    }
};

// Helper function to get stored date from localStorage
function getStoredDate() {
    try {
        const stored = localStorage.getItem('selectedDate');
        if (stored) {
            const date = new Date(stored);
            if (!isNaN(date.getTime())) {
                return date;
            }
        }
    } catch (error) {
        console.error('Error reading stored date:', error);
    }
    return new Date();
}

// Helper function to save date to localStorage
function saveDate(date) {
    try {
        localStorage.setItem('selectedDate', date.toISOString());
    } catch (error) {
        console.error('Error saving date:', error);
    }
}

// Main App component
function App() {
    const [currentDate, setCurrentDate] = useState(getStoredDate());
    const [transactions, setTransactions] = useState([]);
    const [categories, setCategories] = useState([]);
    const [banks, setBanks] = useState([]);
    const [users, setUsers] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [success, setSuccess] = useState(null);
    const [showMonzoAuthModal, setShowMonzoAuthModal] = useState(false);
    const [pendingMonzoImport, setPendingMonzoImport] = useState(null);
    const [activeView, setActiveView] = useState('transactions'); // New state for view switching

    // Update stored date whenever currentDate changes
    useEffect(() => {
        saveDate(currentDate);
    }, [currentDate]);

    // Load initial data
    useEffect(() => {
        const loadInitialData = async () => {
            try {
                const [banksData, usersData, categoriesData] = await Promise.all([
                    API.getBanks(),
                    API.getUsers(),
                    API.getCategories()
                ]);
                console.log('Banks data from API:', banksData);
                console.log('Users data from API:', usersData);
                setBanks(banksData.banks || []);
                setUsers(usersData.users || []);
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

    // Handle Monzo callback from URL
    useEffect(() => {
        const processMonzoCallback = async () => {
            const urlParams = new URLSearchParams(window.location.search);
            const code = urlParams.get('code');
            const state = urlParams.get('state');

            if (code && state) {
                try {
                    // Clean URL immediately
                    window.history.replaceState({}, document.title, window.location.pathname);

                    // Show loading state
                    setError(null);
                    setSuccess(null);

                    // Call backend MonzoCallback
                    console.log('Calling Monzo callback with code and state...');
                    const callbackResult = await API.callMonzoCallback(code, state);

                    if (callbackResult.success) {
                        // Get stored import parameters
                        const storedParams = sessionStorage.getItem('pendingMonzoImport');
                        if (storedParams) {
                            setPendingMonzoImport(JSON.parse(storedParams));
                        }

                        // Show auth modal
                        setShowMonzoAuthModal(true);
                    } else {
                        setError('Failed to process Monzo authorization');
                    }
                } catch (error) {
                    console.error('Error processing Monzo callback:', error);
                    setError('Failed to process Monzo authorization: ' + error.message);
                }
            }
        };

        processMonzoCallback();
    }, []);

    const handleUpload = async (bankId, userId, file) => {
        setError(null);
        setSuccess(null);
        try {
            const result = await API.uploadCSV(bankId, userId, file);

            // Check for errors in response
            if (result.record_error && result.record_error.length > 0) {
                const errorCount = result.record_error.length;
                const firstErrors = result.record_error.slice(0, 3);
                const errorMessages = firstErrors.map(e =>
                    `Row ${e.row_id}: ${e.errors.join(', ')}`
                ).join('; ');

                setError(`Upload completed with ${errorCount} errors. ${errorMessages}${errorCount > 3 ? '...' : ''}`);
            } else if (result.success) {
                setSuccess('Successfully uploaded transactions');
            } else {
                setError('Upload failed');
            }
        } catch (err) {
            setError('Failed to upload CSV file: ' + err.message);
        } finally {
            // ALWAYS reload transactions for current month
            try {
                const data = await API.getTransactions(
                    currentDate.getMonth() + 1,
                    currentDate.getFullYear()
                );
                setTransactions(data.transactions || []);
            } catch (reloadError) {
                console.error('Error reloading transactions:', reloadError);
            }
        }
    };

    const handleTransactionUpdate = async (transactionId, updates) => {
        try {
            const result = await API.updateTransaction(transactionId, updates);

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

    const handleMonzoImport = async (bankId, userId) => {
        setError(null);
        setSuccess(null);
        try {
            console.log('Starting Monzo import with params:', { bankId, userId });
            const month = currentDate.getMonth() + 1;
            const year = currentDate.getFullYear();

            const result = await API.loadMonzoTransactions(month, year, userId, bankId);

            if (result.needsAuth) {
                console.log('Need auth, getting auth URL...');
                const authData = await API.getMonzoAuthURL();

                const authUrl = authData.authUrl || authData.auth_url || authData.url;
                console.log('Auth URL to redirect:', authUrl);

                if (authUrl) {
                    sessionStorage.setItem('pendingMonzoImport', JSON.stringify({
                        bankId,
                        userId,
                        month,
                        year
                    }));
                    window.location.href = authUrl;
                } else {
                    console.error('No auth URL found in response:', authData);
                    setError('Failed to get Monzo authorization URL');
                }
                return;
            }

            // New response format - only success boolean
            if (result.success) {
                console.log('Monzo transactions imported successfully');
                setSuccess('Successfully imported transactions from Monzo');
            } else {
                setError('Failed to import transactions from Monzo');
            }
        } catch (err) {
            console.error('Monzo import error:', err);
            setError('Failed to import from Monzo: ' + err.message);
        } finally {
            // ALWAYS reload transactions for current view
            try {
                const data = await API.getTransactions(
                    currentDate.getMonth() + 1,
                    currentDate.getFullYear()
                );
                setTransactions(data.transactions || []);
            } catch (reloadError) {
                console.error('Error reloading transactions:', reloadError);
            }
        }
    };

    const handleMonzoAuthConfirm = async () => {
        try {
            console.log('Confirming Monzo account access...');
            const accountResult = await API.confirmMonzoAccount();

            if (!accountResult.success) {
                return false;
            }

            console.log('Monzo account confirmed successfully');

            // Close modal
            setShowMonzoAuthModal(false);

            // Now load transactions with stored parameters
            if (pendingMonzoImport) {
                const { bankId, userId, month, year } = pendingMonzoImport;

                try {
                    const result = await API.loadMonzoTransactions(
                        month || currentDate.getMonth() + 1,
                        year || currentDate.getFullYear(),
                        userId,
                        bankId
                    );

                    // New response format - only success boolean
                    if (result.success) {
                        setSuccess('Successfully imported transactions from Monzo');
                    } else {
                        setError('Failed to import transactions from Monzo');
                    }
                } catch (importError) {
                    console.error('Error importing Monzo transactions:', importError);
                    setError('Connected to Monzo successfully, but failed to import transactions: ' + importError.message);
                } finally {
                    // ALWAYS reload transactions for current view, regardless of import result
                    try {
                        const data = await API.getTransactions(
                            currentDate.getMonth() + 1,
                            currentDate.getFullYear()
                        );
                        setTransactions(data.transactions || []);
                    } catch (reloadError) {
                        console.error('Error reloading transactions:', reloadError);
                    }
                }

                // Clear stored parameters
                setPendingMonzoImport(null);
                sessionStorage.removeItem('pendingMonzoImport');
            }

            return true;
        } catch (error) {
            console.error('Error confirming Monzo account:', error);
            throw error;
        }
    };

    const handleMonzoAuthClose = () => {
        setShowMonzoAuthModal(false);
        setPendingMonzoImport(null);
        sessionStorage.removeItem('pendingMonzoImport');
    };

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

            {/* View Tabs */}
            <div className="view-tabs-container">
                <button
                    className={`view-tab ${activeView === 'transactions' ? 'active' : ''}`}
                    onClick={() => setActiveView('transactions')}
                >
                    Transactions
                </button>
                <button
                    className={`view-tab ${activeView === 'charts' ? 'active' : ''}`}
                    onClick={() => setActiveView('charts')}
                >
                    Analytics
                </button>
            </div>

            {activeView === 'transactions' ? (
                <>
                    <UploadTransactions
                        banks={banks}
                        users={users}
                        onUpload={handleUpload}
                        onMonzoImport={handleMonzoImport}
                    />

                    <TransactionsList
                        currentDate={currentDate}
                        onDateChange={setCurrentDate}
                        transactions={transactions}
                        categories={categories}
                        onTransactionUpdate={handleTransactionUpdate}
                        loading={loading}
                    />
                </>
            ) : (
                <ChartsView categories={categories} />
            )}

            {showMonzoAuthModal && (
                <MonzoAuthModal
                    onConfirm={handleMonzoAuthConfirm}
                    onClose={handleMonzoAuthClose}
                    error={null}
                />
            )}
        </div>
    );
}

export default App;