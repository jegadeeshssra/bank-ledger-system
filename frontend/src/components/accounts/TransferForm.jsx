import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAccounts, useTransfer } from '@/hooks/useAccounts'
import { generateIdempotencyKey } from '@/lib/api'

export default function TransferForm({ fromAccountId, fromAccountName, currency, onSuccess }) {
  const [toAccountId, setToAccountId] = useState('')
  const [amount, setAmount] = useState('')
  const [error, setError] = useState('')
  const { data: accounts = [], isLoading } = useAccounts()
  const transfer = useTransfer(fromAccountId, onSuccess)

  const destinationAccounts = accounts.filter((account) => account.id !== fromAccountId)

  const handleSubmit = (event) => {
    event.preventDefault()
    setError('')

    if (!toAccountId.trim()) {
      setError('Please enter a destination account ID.')
      return
    }

    const value = amount.trim()
    const numeric = parseFloat(value)
    if (!value || Number.isNaN(numeric) || numeric <= 0) {
      setError('Please enter a valid positive amount.')
      return
    }

    transfer.mutate({
      data: {
        from_account_id: fromAccountId,
        to_account_id: toAccountId.trim(),
        amount: value,
      },
      idempotencyKey: generateIdempotencyKey(),
    })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-1.5">
        <Label htmlFor="from-account">From account</Label>
        <input
          id="from-account"
          type="text"
          value={fromAccountName}
          disabled
          className="flex h-11 w-full rounded-md border border-input bg-muted/50 px-3 py-2 text-sm text-muted-foreground"
        />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="to-account">To account ID</Label>
        <Input
          id="to-account"
          type="text"
          placeholder="Enter the recipient account UUID"
          value={toAccountId}
          onChange={(event) => setToAccountId(event.target.value)}
          disabled={isLoading || transfer.isPending}
        />
        <p className="text-xs text-muted-foreground">
          Enter the recipient account UUID here, including other user account IDs if needed.
        </p>
      </div>

      {destinationAccounts.length > 0 && (
        <div className="rounded-xl border border-border/70 bg-muted/50 p-3 text-xs text-muted-foreground">
          <p className="font-medium text-foreground text-sm mb-2">Your other accounts</p>
          <div className="space-y-1">
            {destinationAccounts.map((account) => (
              <div key={account.id} className="rounded-md border border-border/60 bg-background/70 px-3 py-2">
                <p className="font-medium text-sm text-foreground">{account.name}</p>
                <p className="text-xs font-mono text-muted-foreground">{account.id}</p>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="space-y-1.5">
        <Label htmlFor="amount">Amount</Label>
        <Input
          id="amount"
          type="number"
          step="0.01"
          min="0"
          placeholder="0.00"
          value={amount}
          onChange={(event) => setAmount(event.target.value)}
          disabled={transfer.isPending}
        />
        <p className="text-xs text-muted-foreground">Amount in {currency ?? 'USD'}</p>
      </div>

      {error && (
        <div className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {error}
        </div>
      )}

      <Button type="submit" className="w-full" disabled={transfer.isPending}>
        {transfer.isPending ? 'Transferring…' : 'Send transfer'}
      </Button>

      {destinationAccounts.length === 0 && (
        <p className="text-sm text-muted-foreground">
          You need at least one additional account to transfer funds.
        </p>
      )}
    </form>
  )
}
