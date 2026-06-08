import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useDeposit, useWithdraw } from '@/hooks/useAccounts'
import { generateIdempotencyKey } from '@/lib/api'

export default function AmountForm({ accountId, mode = 'deposit', currency, onSuccess }) {
  const [amount, setAmount] = useState('')
  const [error, setError] = useState('')

  const action = mode === 'withdraw' ? 'Withdraw' : 'Deposit'
  const mutation = mode === 'withdraw'
    ? useWithdraw(accountId, onSuccess)
    : useDeposit(accountId, onSuccess)

  const handleSubmit = (event) => {
    event.preventDefault()
    setError('')

    const value = amount.trim()
    const numeric = parseFloat(value)
    if (!value || Number.isNaN(numeric) || numeric <= 0) {
      setError('Please enter a valid positive amount.')
      return
    }

    mutation.mutate({ amount: value, idempotencyKey: generateIdempotencyKey() })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
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
          disabled={mutation.isPending}
        />
        <p className="text-xs text-muted-foreground">Amount in {currency ?? 'USD'}</p>
      </div>

      {error && (
        <div className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {error}
        </div>
      )}

      <Button type="submit" className="w-full" disabled={mutation.isPending}>
        {mutation.isPending ? `${action}…` : action}
      </Button>
    </form>
  )
}
