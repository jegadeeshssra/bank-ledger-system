import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { useAccounts } from '@/hooks/useAccounts'
import CreateAccountForm from '@/components/accounts/CreateAccountForm'

export default function DashboardPage() {
  const navigate = useNavigate()
  const { data: accounts, isLoading, isError, error } = useAccounts()
  const [createOpen, setCreateOpen] = useState(false)

  return (
    <div className="space-y-6 max-w-6xl">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-foreground">Accounts</h1>
          <p className="text-sm text-muted-foreground mt-0.5">
            {isLoading ? 'Loading…' : `${accounts?.length ?? 0} account${accounts?.length === 1 ? '' : 's'}`}
          </p>
        </div>
        <Button size="sm" onClick={() => setCreateOpen(true)}>
          <svg className="w-4 h-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
          </svg>
          New account
        </Button>
      </div>

      {/* Error */}
      {isError && (
        <div className="rounded-md bg-destructive/10 border border-destructive/20 px-4 py-3 text-sm text-destructive">
          {error?.message || 'Failed to load accounts.'}
        </div>
      )}

      {/* Loading skeletons */}
      {isLoading && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <Card key={i} className="border-border bg-card">
              <CardHeader className="pb-3">
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-3 w-16 mt-1" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-8 w-24" />
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Empty state */}
      {!isLoading && !isError && accounts?.length === 0 && (
        <div className="flex flex-col items-center justify-center rounded-xl border border-dashed border-border bg-card/50 py-16 text-center">
          <div className="w-12 h-12 rounded-full bg-muted flex items-center justify-center mb-3">
            <svg className="w-6 h-6 text-muted-foreground" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 18.75a60.07 60.07 0 0 1 15.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 0 1 3 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 0 0-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 0 1-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 0 0 3 15h-.75M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Zm3 0h1.5a.75.75 0 0 0 0-1.5H18M9 10.5H7.5a.75.75 0 0 0 0 1.5H9" />
            </svg>
          </div>
          <p className="text-sm font-medium text-foreground">No accounts yet</p>
          <p className="text-xs text-muted-foreground mt-1 mb-4">Create your first account to get started</p>
          <Button size="sm" onClick={() => setCreateOpen(true)}>Create account</Button>
        </div>
      )}

      {/* Account grid */}
      {!isLoading && accounts && accounts.length > 0 && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {accounts.map((acc) => (
            <AccountCard key={acc.id} account={acc} onClick={() => navigate(`/accounts/${acc.id}`)} />
          ))}
        </div>
      )}

      {/* Create Account Dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-xl lg:max-w-2xl">
          <DialogHeader>
            <DialogTitle>New account</DialogTitle>
            <DialogDescription>Create a new ledger account</DialogDescription>
          </DialogHeader>
          <CreateAccountForm onSuccess={() => setCreateOpen(false)} />
        </DialogContent>
      </Dialog>
    </div>
  )
}

function AccountCard({ account, onClick }) {
  const balance = parseFloat(account.balance ?? '0')
  const isNegative = balance < 0

  return (
    <Card
      className="border-border bg-card hover:border-primary/40 hover:bg-card/80 transition-all duration-150 cursor-pointer group"
      onClick={onClick}
    >
      <CardHeader className="pb-2">
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0">
            <p className="font-medium text-foreground truncate leading-none">{account.name}</p>
            <p className="text-xs text-muted-foreground mt-1 font-mono">{account.id.slice(0, 8)}…</p>
          </div>
          <div className="flex gap-1 shrink-0">
            <Badge variant="secondary" className="text-xs px-1.5 py-0 h-5 font-mono">
              {account.currency}
            </Badge>
            {account.is_system && (
              <Badge variant="outline" className="text-xs px-1.5 py-0 h-5 border-primary/40 text-primary">
                SYS
              </Badge>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        <div className="flex items-end justify-between">
          <p className={`text-2xl font-semibold font-mono tabular-nums ${isNegative ? 'text-destructive' : 'text-foreground'}`}>
            {isNegative ? '–' : ''}{Math.abs(balance).toFixed(2)}
          </p>
          <svg className="w-4 h-4 text-muted-foreground group-hover:text-primary transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" />
          </svg>
        </div>
        <p className="text-xs text-muted-foreground mt-1">
          Created {new Date(account.created_at).toLocaleDateString()}
        </p>
      </CardContent>
    </Card>
  )
}
