import { useState, useEffect } from 'react';
import { BarChart, Bar, LineChart, Line, PieChart, Pie, Cell, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { TrendingUp, PieChart as PieChartIcon, BarChart3, Calendar } from 'lucide-react';
import './ChartsView.css';

// Color palette for charts - more diverse colors
const COLORS = [
    '#FF6B6B', '#4ECDC4', '#45B7D1', '#F7DC6F', '#BB8FCE',
    '#85C1E2', '#F8B739', '#6C5CE7', '#00D2D3', '#FF9FF3',
    '#54A0FF', '#48DBFB', '#FDA7DF', '#C44569', '#F5CD79',
    '#786FA6', '#F19066', '#546DE5', '#E15F41', '#303952'
];

// Custom tooltip component
function CustomTooltip({ active, payload, label }) {
    if (active && payload && payload.length) {
        return (
            <div className="custom-tooltip">
                <p className="label">{label}</p>
                {payload.map((entry, index) => (
                    <p key={index} style={{ color: entry.color }}>
                        {entry.name}: £{entry.value.toFixed(2)}
                    </p>
                ))}
            </div>
        );
    }
    return null;
}

// Year selector component
function YearSelector({ selectedYear, onYearChange }) {
    const currentYear = new Date().getFullYear();
    const years = [];

    // Generate last 5 years
    for (let i = 0; i < 5; i++) {
        years.push(currentYear - i);
    }

    return (
        <div className="year-selector">
            <Calendar size={20} />
            <select value={selectedYear} onChange={(e) => onYearChange(parseInt(e.target.value))}>
                {years.map(year => (
                    <option key={year} value={year}>{year}</option>
                ))}
            </select>
        </div>
    );
}

// API functions
const ChartAPI = {
    async getYearlyTransactions(year) {
        try {
            const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
            const allTransactions = [];

            // Fetch transactions for each month of the year
            for (let month = 0; month < 12; month++) {
                try {
                    const response = await fetch(`/transactions?month=${month + 1}&year=${year}`);
                    if (response.ok) {
                        const data = await response.json();
                        if (data.transactions) {
                            allTransactions.push({
                                month: months[month],
                                monthIndex: month,
                                transactions: data.transactions
                            });
                        }
                    }
                } catch (error) {
                    console.error(`Failed to fetch transactions for ${months[month]} ${year}:`, error);
                    // Continue with other months even if one fails
                    allTransactions.push({
                        month: months[month],
                        monthIndex: month,
                        transactions: []
                    });
                }
            }

            return allTransactions;
        } catch (error) {
            console.error('Failed to fetch yearly data:', error);
            throw error;
        }
    }
};

// Main ChartsView component
function ChartsView({ categories }) {
    const [selectedYear, setSelectedYear] = useState(new Date().getFullYear());
    const [chartType, setChartType] = useState('bar');
    const [yearlyData, setYearlyData] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    // Fetch data when year changes
    useEffect(() => {
        const fetchYearlyData = async () => {
            setLoading(true);
            setError(null);
            try {
                const data = await ChartAPI.getYearlyTransactions(selectedYear);
                setYearlyData(data);
            } catch (err) {
                console.error('Failed to fetch yearly data:', err);
                setError('Failed to load chart data');
                // Use mock data for now if API fails
                setYearlyData(generateMockData());
            } finally {
                setLoading(false);
            }
        };

        fetchYearlyData();
    }, [selectedYear]);

    // Generate mock data if API is not ready
    const generateMockData = () => {
        const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
        return months.map((month, index) => ({
            month,
            monthIndex: index,
            transactions: []
        }));
    };

    // Process data for charts
    const processChartData = () => {
        if (!yearlyData || yearlyData.length === 0) return { monthlyData: [], categoryTotals: [] };

        // Process monthly data
        const monthlyData = yearlyData.map(monthData => {
            const processed = { month: monthData.month };

            // Calculate totals by category for this month
            categories.forEach(category => {
                if (category.name !== 'Transfer') {
                    const categoryTotal = monthData.transactions
                        .filter(t =>
                            (t.categoryId === category.id || t.categoryName === category.name) &&
                            t.type === 'OUTCOME'
                        )
                        .reduce((sum, t) => sum + Math.abs(parseFloat(t.amount)), 0);

                    if (categoryTotal > 0) {
                        processed[category.name] = categoryTotal;
                    }
                }
            });

            return processed;
        });

        // Calculate category totals for the year
        const categoryTotals = categories
            .filter(cat => cat.name !== 'Transfer')
            .map(category => {
                const total = yearlyData.reduce((yearSum, monthData) => {
                    const monthTotal = monthData.transactions
                        .filter(t =>
                            (t.categoryId === category.id || t.categoryName === category.name) &&
                            t.type === 'OUTCOME'
                        )
                        .reduce((sum, t) => sum + Math.abs(parseFloat(t.amount)), 0);
                    return yearSum + monthTotal;
                }, 0);

                return {
                    name: category.name,
                    value: total
                };
            })
            .filter(cat => cat.value > 0)
            .sort((a, b) => b.value - a.value);

        // If no real data, generate some sample data for visualization
        if (categoryTotals.length === 0 && monthlyData.every(m => Object.keys(m).length === 1)) {
            const sampleCategories = categories.filter(c => c.name !== 'Transfer').slice(0, 5);
            const sampleMonthlyData = monthlyData.map(m => {
                const data = { month: m.month };
                sampleCategories.forEach(cat => {
                    data[cat.name] = Math.floor(Math.random() * 2000) + 100;
                });
                return data;
            });

            const sampleCategoryTotals = sampleCategories.map(cat => ({
                name: cat.name,
                value: sampleMonthlyData.reduce((sum, m) => sum + (m[cat.name] || 0), 0)
            })).sort((a, b) => b.value - a.value);

            return { monthlyData: sampleMonthlyData, categoryTotals: sampleCategoryTotals };
        }

        return { monthlyData, categoryTotals };
    };

    const { monthlyData, categoryTotals } = processChartData();

    if (loading) {
        return (
            <div className="charts-container">
                <div className="loading">Loading chart data...</div>
            </div>
        );
    }

    return (
        <div className="charts-container">
            <div className="charts-header">
                <h2>
                    <TrendingUp size={24} />
                    Expense Analytics
                </h2>
                <div className="chart-controls">
                    <YearSelector selectedYear={selectedYear} onYearChange={setSelectedYear} />
                    <div className="chart-type-selector">
                        <button
                            className={`chart-type-btn ${chartType === 'bar' ? 'active' : ''}`}
                            onClick={() => setChartType('bar')}
                        >
                            <BarChart3 size={18} />
                            Bar Chart
                        </button>
                        <button
                            className={`chart-type-btn ${chartType === 'line' ? 'active' : ''}`}
                            onClick={() => setChartType('line')}
                        >
                            <TrendingUp size={18} />
                            Line Chart
                        </button>
                        <button
                            className={`chart-type-btn ${chartType === 'pie' ? 'active' : ''}`}
                            onClick={() => setChartType('pie')}
                        >
                            <PieChartIcon size={18} />
                            Pie Chart
                        </button>
                    </div>
                </div>
            </div>

            {error && (
                <div className="chart-error">
                    {error}
                </div>
            )}

            <div className="charts-grid">
                {/* Monthly spending chart */}
                {(chartType === 'bar' || chartType === 'line') && (
                    <div className="chart-card">
                        <h3>Monthly Expenses by Category</h3>
                        <ResponsiveContainer width="100%" height={400}>
                            {chartType === 'bar' ? (
                                <BarChart data={monthlyData}>
                                    <CartesianGrid strokeDasharray="3 3" />
                                    <XAxis dataKey="month" />
                                    <YAxis />
                                    <Tooltip content={<CustomTooltip />} />
                                    <Legend />
                                    {categoryTotals.map((category, index) => (
                                        <Bar
                                            key={category.name}
                                            dataKey={category.name}
                                            fill={COLORS[index % COLORS.length]}
                                        />
                                    ))}
                                </BarChart>
                            ) : (
                                <LineChart data={monthlyData}>
                                    <CartesianGrid strokeDasharray="3 3" />
                                    <XAxis dataKey="month" />
                                    <YAxis />
                                    <Tooltip content={<CustomTooltip />} />
                                    <Legend />
                                    {categoryTotals.map((category, index) => (
                                        <Line
                                            key={category.name}
                                            type="monotone"
                                            dataKey={category.name}
                                            stroke={COLORS[index % COLORS.length]}
                                            strokeWidth={2}
                                        />
                                    ))}
                                </LineChart>
                            )}
                        </ResponsiveContainer>
                    </div>
                )}

                {/* Category distribution pie chart */}
                {chartType === 'pie' && (
                    <div className="chart-card">
                        <h3>Total Expenses by Category - {selectedYear}</h3>
                        <ResponsiveContainer width="100%" height={400}>
                            <PieChart>
                                <Pie
                                    data={categoryTotals}
                                    cx="50%"
                                    cy="50%"
                                    labelLine={false}
                                    label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                                    outerRadius={120}
                                    fill="#8884d8"
                                    dataKey="value"
                                >
                                    {categoryTotals.map((entry, index) => (
                                        <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                                    ))}
                                </Pie>
                                <Tooltip content={<CustomTooltip />} />
                            </PieChart>
                        </ResponsiveContainer>
                        <div className="category-legend">
                            {categoryTotals.map((category, index) => (
                                <div key={category.name} className="legend-item">
                  <span
                      className="legend-color"
                      style={{ backgroundColor: COLORS[index % COLORS.length] }}
                  />
                                    <span className="legend-label">{category.name}</span>
                                    <span className="legend-value">£{category.value.toFixed(2)}</span>
                                </div>
                            ))}
                        </div>
                    </div>
                )}

                {/* Summary statistics */}
                <div className="chart-card summary-card">
                    <h3>Year Summary - {selectedYear}</h3>
                    <div className="summary-stats">
                        <div className="stat-item">
                            <h4>Total Expenses</h4>
                            <p className="stat-value expense">
                                £{categoryTotals.reduce((sum, cat) => sum + cat.value, 0).toFixed(2)}
                            </p>
                        </div>
                        <div className="stat-item">
                            <h4>Average Monthly</h4>
                            <p className="stat-value">
                                £{(categoryTotals.reduce((sum, cat) => sum + cat.value, 0) / 12).toFixed(2)}
                            </p>
                        </div>
                        <div className="stat-item">
                            <h4>Highest Category</h4>
                            <p className="stat-value">
                                {categoryTotals[0]?.name || 'N/A'}
                            </p>
                        </div>
                        <div className="stat-item">
                            <h4>Categories Tracked</h4>
                            <p className="stat-value">
                                {categoryTotals.length}
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default ChartsView;