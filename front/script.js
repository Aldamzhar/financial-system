document.getElementById('create-account-form').addEventListener('submit', function(e) {
    e.preventDefault();
    const name = document.getElementById('name').value;
    const balance = parseFloat(document.getElementById('balance').value);
    fetch('/accounts', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ name, balance }),
    })
    .then(response => response.json())
    .then(data => {
        console.log(data);
        fetchAccounts();
    })
    .catch(error => console.error('Error:', error));
});

function fetchAccounts() {
    fetch('/accounts')
        .then(response => response.json())
        .then(accounts => {
            accounts.sort((a, b) => b.balance - a.balance);
            const accountsContainer = document.getElementById('accounts-container');
            accountsContainer.innerHTML = '';

            accounts.forEach(account => {
                const accountDiv = document.createElement('div');
                accountDiv.className = 'account';
                accountDiv.setAttribute('data-account-id', account.id);
                accountDiv.innerHTML = `
                    Account ID: <strong>${account.id}</strong> <br>
                    Name: <strong>${account.name}</strong> <br>
                    <span class="balance">Balance: <strong>${account.balance}</strong></span> <br>
                `;
                accountsContainer.appendChild(accountDiv);

                const transactionsDiv = document.createElement('div');
                transactionsDiv.id = 'transactions-' + account.id;
                accountDiv.appendChild(transactionsDiv);

                fetchTransactions(account.id, transactionsDiv.id);
            });
        })
        .catch(error => console.error('Error fetching accounts:', error));
}


function fetchTransactions(accountId, transactionsDivId) {
    fetch(`/accounts/${accountId}/transactions`)
        .then(response => response.json())
        .then(transactions => {
            const transactionsDiv = document.getElementById(transactionsDivId);
            const table = document.createElement('table');
            table.innerHTML = '<tr><th>ID</th><th>Value</th><th>Group Type</th><th>Transfer To</th><th>From</th><th>Date</th></tr>';

            transactions.forEach(tr => {
                const row = table.insertRow();
                row.insertCell(0).innerText = tr.id;
                row.insertCell(1).innerText = tr.value;
                row.insertCell(2).innerText = tr.group_type;
                row.insertCell(3).innerText = tr.account2_id || 'N/A';
                row.insertCell(4).innerText = tr.account_id; 
                row.insertCell(5).innerText = new Date(tr.transaction_date).toLocaleString();
            });
            transactionsDiv.appendChild(table);
        })
        .catch(error => console.error('Error fetching transactions for account ID ' + accountId + ':', error));
}

document.getElementById('create-transaction-form').addEventListener('submit', function(e) {
    e.preventDefault();
    const accountId = parseInt(document.getElementById('account-id').value, 10);
    const account2Id = document.getElementById('account2-id').value ? parseInt(document.getElementById('account2-id').value, 10) : null;
    const value = parseFloat(document.getElementById('value').value);
    const groupType = document.getElementById('group-type').value;

    const accountDiv = document.querySelector(`div.account[data-account-id="${accountId}"]`);
    const balanceElement = accountDiv ? accountDiv.querySelector('.balance strong') : null;
    let currentBalance = balanceElement ? parseFloat(balanceElement.innerText.replace(/,/g, '')) : null;

    if (currentBalance !== null) {
        if (groupType === 'transfer' && (currentBalance < value || value <= 0)) {
            console.error('Insufficient funds');
            alert('Insufficient funds');
            return;
        }

        if (groupType === 'outcome' && currentBalance < value) {
            console.error('Insufficient funds.');
            alert('Insufficient funds.');
            return;
        }
    }


    const transactionDetails = {
        account_id: accountId,
        account2_id: account2Id,
        value: value,
        group_type: groupType,
        transaction_date: new Date().toISOString()
    };

    fetch('/transactions', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(transactionDetails),
    })
    .then(response => response.json())
    .then(data => {
        console.log(data);
        updateAccountBalances(accountId, account2Id, value, groupType);
        addTransactionToTable(data);
    })
    .catch(error => console.error('Error:', error));
});

function createTransactionTable(container) {
    const table = document.createElement('table');
    container.appendChild(table);
    table.innerHTML = '<tr><th>ID</th><th>Value</th><th>Group Type</th><th>Transfer To</th><th>From</th><th>Date</th></tr>';
    return table;
}

function addTransactionToTable(transaction) {
    const transactionsDiv = document.getElementById('transactions-' + transaction.account_id);
    if (transactionsDiv) {
        const table = transactionsDiv.querySelector('table') || createTransactionTable(transactionsDiv);
        const row = table.insertRow();
        row.insertCell(0).innerText = transaction.id;
        row.insertCell(1).innerText = transaction.value;
        row.insertCell(2).innerText = transaction.group_type;
        row.insertCell(3).innerText = transaction.account2_id || 'N/A';
        row.insertCell(4).innerText = transaction.account_id;
        row.insertCell(5).innerText = new Date(transaction.transaction_date).toLocaleString();
    }
}


function updateAccountBalances(accountId, account2Id, value, groupType) {
    adjustBalance(accountId, groupType === 'income' ? value : -value);
    if (groupType === 'transfer' && account2Id) {
        adjustBalance(account2Id, value);
    }
}

function adjustBalance(accountId, valueChange) {
    const accountDiv = document.querySelector(`div.account[data-account-id="${accountId}"]`);
    if (accountDiv) {
        const balanceElement = accountDiv.querySelector('.balance strong');
        let currentBalance = parseFloat(balanceElement.innerText.replace(/,/g, ''));
        let newBalance = currentBalance + valueChange;
        balanceElement.innerText = newBalance
    }
}


window.onload = fetchAccounts;


