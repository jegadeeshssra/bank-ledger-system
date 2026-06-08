import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useCreateAccount } from '@/hooks/useAccounts'

const DEFAULT_CURRENCY = 'USD'
const currencies = ['USD', 'EUR', 'GBP', 'INR']

export default function CreateAccountForm({ onSuccess }) {
  const [name, setName] = useState('')
  const [currency, setCurrency] = useState(DEFAULT_CURRENCY)
  const [isSystem, setIsSystem] = useState(false)
  const [error, setError] = useState('')

  const createAccount = useCreateAccount(() => {
    setName('')
    setCurrency(DEFAULT_CURRENCY)
    setIsSystem(false)
    onSuccess?.()
  })

  const handleSubmit = (event) => {
    event.preventDefault()
    setError('')

    if (!name.trim()) {
      setError('Account name is required.')
      return
    }
    if (currency.trim().length !== 3) {
      setError('Currency code must be 3 letters.')
      return
    }

    createAccount.mutate({
      name: name.trim(),
      currency: currency.trim().toUpperCase(),
      is_system: isSystem,
    })
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-1.5">
        <Label htmlFor="account-name">Account name</Label>
        <Input
          id="account-name"
          type="text"
          placeholder="Primary checking"
          value={name}
          onChange={(event) => setName(event.target.value)}
          disabled={createAccount.isPending}
        />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="currency">Currency</Label>
        <select
          id="currency"
          className="flex h-11 w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50"
          value={currency}
          onChange={(event) => setCurrency(event.target.value)}
          disabled={createAccount.isPending}
        >
          {currencies.map((item) => (
            <option key={item} value={item}>
              {item}
            </option>
          ))}
        </select>
      </div>

      <div className="flex items-center gap-2">
        <input
          id="is-system"
          type="checkbox"
          checked={isSystem}
          onChange={(event) => setIsSystem(event.target.checked)}
          disabled={createAccount.isPending}
          className="h-4 w-4 rounded border-input text-primary focus:ring-primary"
        />
        <Label htmlFor="is-system" className="mb-0 text-sm">
          System account
        </Label>
      </div>

      {error && (
        <div className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          {error}
        </div>
      )}

      <Button type="submit" className="w-full" disabled={createAccount.isPending}>
        {createAccount.isPending ? 'Creating account…' : 'Create account'}
      </Button>
    </form>
  )
}
