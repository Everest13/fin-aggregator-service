import { useState, useEffect, useRef } from 'react';
import { Calendar, ChevronLeft, ChevronRight, ChevronDown } from 'lucide-react';
import './TransactionsList.css';

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
    const [viewMode, setViewMode] = useState('categories');

    if (transactions.length === 0) {
        return (
            <div className="empty-state">
                <Calendar size={48} />
                <h3>No transactions found</h3>
                <p>Upload a CSV file or import from your bank to see transactions for this month</p>
            </div>
        );
    }

    // Helper function to check if category is Transfer
    const isTransferCategory = (categoryId, categoryName) => {
        const category = categories.find(c => c.id === parseInt(categoryId) || c.id === categoryId);
        return category?.name === 'Transfer' || categoryName === 'Transfer';
    };

    // Calculate totals excluding Transfer category
    const totals = transactions.reduce((acc, transaction) => {
        const amount = Math.abs(parseFloat(transaction.amount));

        // Skip Transfer category transactions for totals
        if (isTransferCategory(transaction.categoryId, transaction.categoryName)) {
            return acc;
        }

        if (transaction.type === 'INCOME') {
            acc.income += amount;
        } else if (transaction.type === 'OUTCOME') {
            acc.outcome += amount;
        }
        return acc;
    }, { income: 0, outcome: 0 });

    const transactionsByCategory = transactions.reduce((acc, transaction) => {
        const categoryId = transaction.categoryId ? String(transaction.categoryId) : 'uncategorized';

        if (!acc[categoryId]) {
            acc[categoryId] = {
                transactions: [],
                income: 0,
                outcome: 0
            };
        }

        acc[categoryId].transactions.push(transaction);

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
        return `${day}\u00A0${month}`;
    };

    const formatAmount = (amount, type) => {
        const value = Math.abs(parseFloat(amount));
        const formatted = value.toFixed(2);

        if (type === 'INCOME') {
            return `+${formatted}`;
        } else if (type === 'OUTCOME') {
            return `-${formatted}`;
        } else {
            return formatted;
        }
    };

    const getCategoryName = (categoryId) => {
        if (categoryId === 'uncategorized') return 'Uncategorized';

        const category = categories.find(c => c.id === parseInt(categoryId) || c.id === categoryId);

        if (category) {
            return category.name;
        }

        return `Category ${categoryId}`;
    };

    return (
        <div>
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

            {viewMode === 'categories' && (
                <div className="categories-container">
                    {Object.entries(transactionsByCategory)
                        .sort(([, a], [, b]) => b.outcome - a.outcome)
                        .map(([categoryId, data]) => {
                            const categoryName = getCategoryName(categoryId);
                            const isTransfer = categoryName === 'Transfer';

                            return (
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
                                            <h3>{categoryName}</h3>
                                            <span className="transaction-count">({data.transactions.length})</span>
                                        </div>
                                        <div className="category-summary">
                                            {data.income > 0 && (
                                                <span className="summary-income">
                                                    {isTransfer ? 'In: ' : 'Income: '}+{data.income.toFixed(2)}
                                                </span>
                                            )}
                                            {data.outcome > 0 && (
                                                <span className="summary-outcome">
                                                    {isTransfer ? 'Out: ' : 'Expenses: '}-{data.outcome.toFixed(2)}
                                                </span>
                                            )}
                                            {isTransfer && (
                                                <span className={data.income - data.outcome >= 0 ? 'summary-income' : 'summary-outcome'}>
                                                    Net: {(data.income - data.outcome).toFixed(2)}
                                                </span>
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
                            );
                        })}
                </div>
            )}

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

// Main TransactionsList component
function TransactionsList({
                              currentDate,
                              onDateChange,
                              transactions,
                              categories,
                              onTransactionUpdate,
                              loading
                          }) {
    return (
        <>
            <MonthNavigation
                currentDate={currentDate}
                onDateChange={onDateChange}
            />

            <div className="transactions-section">
                {loading ? (
                    <div className="loading">Loading transactions...</div>
                ) : (
                    <TransactionsTable
                        transactions={transactions}
                        categories={categories}
                        onTransactionUpdate={onTransactionUpdate}
                    />
                )}
            </div>
        </>
    );
}

export default TransactionsList;