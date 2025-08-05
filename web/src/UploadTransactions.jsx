import { useState, useEffect } from 'react';
import { Upload, FileText, Cloud, AlertCircle } from 'lucide-react';
import './UploadTransactions.css';

// Import type enum mapping (from protobuf)
const BankImportType = {
    UNDEFINED: 0,
    CSV: 1,
    API: 2
};

// Import method configurations
const IMPORT_CONFIGS = {
    CSV: {
        icon: FileText,
        buttonText: 'Upload CSV',
        loadingText: 'Uploading file...',
        helpText: 'Select a CSV file with bank transactions',
        requiresFile: true
    },
    API: {
        icon: Cloud,
        buttonText: 'Import from API',
        loadingText: 'Connecting to bank API...',
        helpText: 'Import transactions directly from your bank account',
        requiresFile: false
    }
};

// Import method renderer component
function ImportMethodRenderer({ importType, bankName, file, onFileChange, onSubmit, isUploading }) {
    const configKey = importType === BankImportType.API ? 'API' : 'CSV';
    const config = IMPORT_CONFIGS[configKey];
    const Icon = config.icon;

    let buttonText = config.buttonText;
    let loadingText = config.loadingText;
    if (importType === BankImportType.API && bankName) {
        buttonText = `Import from ${bankName}`;
        loadingText = `Connecting to ${bankName}...`;
    }

    if (config.requiresFile) {
        return (
            <>
                <div className="form-group">
                    <label htmlFor="file">CSV File</label>
                    <div className="file-input-wrapper">
                        <input
                            id="file"
                            type="file"
                            accept=".csv"
                            onChange={(e) => onFileChange(e.target.files[0])}
                            required
                        />
                        <div className="file-input-help">
                            <Icon size={16} />
                            <span>{config.helpText}</span>
                        </div>
                    </div>
                </div>
                <button
                    onClick={onSubmit}
                    className="upload-btn"
                    disabled={isUploading || !file}
                >
                    <Icon size={20} />
                    {isUploading ? loadingText : buttonText}
                </button>
            </>
        );
    }

    return (
        <button
            onClick={onSubmit}
            className="upload-btn import-api-btn"
            disabled={isUploading}
        >
            <Icon size={20} />
            {isUploading ? loadingText : buttonText}
        </button>
    );
}

// Upload section component
function UploadTransactions({ banks, users, onUpload, onMonzoImport }) {
    const [selectedBank, setSelectedBank] = useState('');
    const [selectedUser, setSelectedUser] = useState('');
    const [file, setFile] = useState(null);
    const [isUploading, setIsUploading] = useState(false);

    const selectedUserObject = users.find(user => String(user.id) === String(selectedUser));

    const availableBanks = selectedUserObject?.banks && selectedUserObject.banks.length > 0
        ? banks.filter(bank => {
            return selectedUserObject.banks.some(userBankId =>
                String(userBankId) === String(bank.id)
            );
        })
        : [];

    const selectedBankObject = banks.find(bank => String(bank.id) === String(selectedBank));
    const importType = selectedBankObject?.importMethod === 'API' ? BankImportType.API : BankImportType.CSV;

    useEffect(() => {
        setSelectedBank('');
        setFile(null);
    }, [selectedUser]);

    useEffect(() => {
        if (importType !== BankImportType.CSV) {
            setFile(null);
        }
    }, [importType]);

    const handleSubmit = async () => {
        if (!selectedBank || !selectedUser) {
            alert('Please select both user and bank');
            return;
        }

        setIsUploading(true);
        try {
            if (importType === BankImportType.CSV) {
                if (!file) {
                    alert('Please select a CSV file');
                    return;
                }
                await onUpload(selectedBank, selectedUser, file);
                setFile(null);
            } else if (importType === BankImportType.API) {
                await onMonzoImport(selectedBank, selectedUser);
            }
        } catch (error) {
            console.error('Import failed:', error);
        } finally {
            setIsUploading(false);
        }
    };

    return (
        <div className="upload-section">
            <h2>
                <Upload size={24} />
                Upload Transactions
            </h2>
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
                        {users.map(user => (
                            <option key={user.id} value={user.id}>
                                {user.name}
                                {user.banks && user.banks.length > 0 && ` (${user.banks.length} banks)`}
                            </option>
                        ))}
                    </select>
                </div>

                <div className="form-group">
                    <label htmlFor="bank">Bank</label>
                    <div className="bank-select-wrapper">
                        <select
                            id="bank"
                            value={selectedBank}
                            onChange={(e) => setSelectedBank(e.target.value)}
                            disabled={!selectedUser}
                            required
                        >
                            <option value="">
                                {!selectedUser
                                    ? "Select a user first"
                                    : availableBanks.length === 0
                                        ? "No banks available for this user"
                                        : "Select a bank"}
                            </option>
                            {availableBanks.map(bank => (
                                <option key={bank.id} value={bank.id}>
                                    {bank.name}
                                </option>
                            ))}
                        </select>
                        {selectedBankObject && (
                            <div className="bank-import-type">
                                {(() => {
                                    const configKey = importType === BankImportType.API ? 'API' : 'CSV';
                                    const config = IMPORT_CONFIGS[configKey];
                                    const Icon = config?.icon || AlertCircle;
                                    return <Icon size={16} />;
                                })()}
                                <span>
                                    {importType === BankImportType.API
                                        ? `API Import`
                                        : 'CSV Upload'}
                                </span>
                            </div>
                        )}
                    </div>
                </div>

                {selectedBank && selectedUser && (
                    <ImportMethodRenderer
                        importType={importType}
                        bankName={selectedBankObject?.name}
                        file={file}
                        onFileChange={setFile}
                        onSubmit={handleSubmit}
                        isUploading={isUploading}
                    />
                )}
            </div>
        </div>
    );
}

export default UploadTransactions;