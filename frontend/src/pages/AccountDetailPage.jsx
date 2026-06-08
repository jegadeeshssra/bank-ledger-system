import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { Separator } from '@/components/ui/separator'
import {
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription,
} from '@/components/ui/dialog'
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { useAccount, useDeleteAccount, useReconcile } from '@/hooks/useAccounts'
import AmountForm from '@/components/accounts/AmountForm'
import TransferForm from '@/components/accounts/TransferForm'

export default function AccountDetailPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const { data: account, isLoading, isError, error } = useAccount(id)

  const [depositOpen, setDepositOpen] = useState(false)
  const [withdrawOpen, setWithdrawOpen] = useState(false)
  const [transferOpen, setTransferOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)

  const deleteAccount = useDeleteAccount()
  const reconcile = useReconcile(id, () => {})

  const handleDelete = () => {
    deleteAccount.mutate(id, {
      onSuccess: () => navigate('/'),
    })
  }

  if (isLoading) return <AccountDetailSkeleton />

  if (isError) return (
    <div className="space-y-4">
      <BackButton onClick={() => navigate('/')} />
      <div className="rounded-md bg-destructive/10 border border-destructive/20 px-4 py-3 text-sm text-destructive">
        {error?.message || 'Failed to load account.'}
      </div>
    </div>
  )

  if (!account) return null

  const balance = parseFloat(account.balance ?? '0')
  const isNegative = balance < 0

  return (
    <div className="space-y-6 max-w-5xl">
      <BackButton onClick={() => navigate('/')} />

      {/* Account header */}
      <div className="rounded-xl border border-border bg-card p-6">
        <div className="flex items-start justify-between gap-4 flex-wrap">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <h1 className="text-xl font-semibold text-foreground">{account.name}</h1>
              <Badge variant="secondary" className="font-mono">{account.currency}</Badge>
              {account.is_system && (
                <Badge variant="outline" className="border-primary/40 text-primary">System</Badge>
              )}
            </div>
            <p className="text-xs text-muted-foreground font-mono">{account.id}</p>
            <p className="text-xs text-muted-foreground mt-1">
              Created {new Date(account.created_at).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })}
            </p>
          </div>
          <div className="text-right">
            <p className="text-xs text-muted-foreground uppercase tracking-wider mb-0.5">Balance</p>
            <p className={`text-3xl font-bold font-mono tabular-nums ${isNegative ? 'text-destructive' : 'text-foreground'}`}>
              {isNegative ? '–' : ''}{Math.abs(balance).toFixed(2)}
            </p>
            <p className="text-xs text-muted-foreground">{account.currency}</p>
          </div>
        </div>
      </div>

      {/* Actions */}
      <div>
        <p className="text-xs text-muted-foreground uppercase tracking-wider mb-3 font-medium">Actions</p>
        <div className="flex flex-wrap gap-2">
          <ActionButton
            label="Deposit"
            icon={<DepositIcon />}
            variant="default"
            onClick={() => setDepositOpen(true)}
          />
          <ActionButton
            label="Withdraw"
            icon={<WithdrawIcon />}
            variant="outline"
            onClick={() => setWithdrawOpen(true)}
          />
          <ActionButton
            label="Transfer"
            icon={<TransferIcon />}
            variant="outline"
            onClick={() => setTransferOpen(true)}
          />
          <ActionButton
            label="Reconcile"
            icon={<ReconcileIcon />}
            variant="outline"
            onClick={() => reconcile.mutate()}
            loading={reconcile.isPending}
          />
          <ActionButton
            label="Entries"
            icon={<ListIcon />}
            variant="outline"
            onClick={() => navigate(`/accounts/${id}/entries`)}
          />
        </div>
      </div>

      <Separator />

      {/* Danger zone */}
      <div>
        <p className="text-xs text-muted-foreground uppercase tracking-wider mb-3 font-medium">Danger zone</p>
        <Button
          variant="outline"
          size="sm"
          className="border-destructive/40 text-destructive hover:bg-destructive/10 hover:border-destructive"
          onClick={() => setDeleteOpen(true)}
        >
          <svg className="w-4 h-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}>
            <path strokeLinecap="round" strokeLinejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" />
          </svg>
          Delete account
        </Button>
      </div>

      {/* Deposit Dialog */}
      <Dialog open={depositOpen} onOpenChange={setDepositOpen}>
        <DialogContent className="sm:max-w-xl lg:max-w-2xl">
          <DialogHeader>
            <DialogTitle>Deposit funds</DialogTitle>
            <DialogDescription>Add funds to <span className="text-foreground font-medium">{account.name}</span></DialogDescription>
          </DialogHeader>
          <AmountForm
            accountId={id}
            mode="deposit"
            currency={account.currency}
            onSuccess={() => setDepositOpen(false)}
          />
        </DialogContent>
      </Dialog>

      {/* Withdraw Dialog */}
      <Dialog open={withdrawOpen} onOpenChange={setWithdrawOpen}>
        <DialogContent className="sm:max-w-xl lg:max-w-2xl">
          <DialogHeader>
            <DialogTitle>Withdraw funds</DialogTitle>
            <DialogDescription>Withdraw from <span className="text-foreground font-medium">{account.name}</span></DialogDescription>
          </DialogHeader>
          <AmountForm
            accountId={id}
            mode="withdraw"
            currency={account.currency}
            onSuccess={() => setWithdrawOpen(false)}
          />
        </DialogContent>
      </Dialog>

      {/* Transfer Dialog */}
      <Dialog open={transferOpen} onOpenChange={setTransferOpen}>
        <DialogContent className="sm:max-w-xl lg:max-w-2xl">
          <DialogHeader>
            <DialogTitle>Transfer funds</DialogTitle>
            <DialogDescription>Transfer from <span className="text-foreground font-medium">{account.name}</span></DialogDescription>
          </DialogHeader>
          <TransferForm
            fromAccountId={id}
            fromAccountName={account.name}
            currency={account.currency}
            onSuccess={() => setTransferOpen(false)}
          />
        </DialogContent>
      </Dialog>

      {/* Delete AlertDialog */}
      <AlertDialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete account?</AlertDialogTitle>
            <AlertDialogDescription>
              This will permanently delete <span className="text-foreground font-medium">{account.name}</span> and all its entries.
              This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
              onClick={handleDelete}
              disabled={deleteAccount.isPending}
            >
              {deleteAccount.isPending ? 'Deleting…' : 'Delete'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

// ── Sub-components ────────────────────────────────────────────────────

function BackButton({ onClick }) {
  return (
    <button
      onClick={onClick}
      className="flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground transition-colors"
    >
      <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
      </svg>
      Back to accounts
    </button>
  )
}

function ActionButton({ label, icon, variant, onClick, loading }) {
  return (
    <Button variant={variant} size="sm" onClick={onClick} disabled={loading}>
      {loading ? (
        <svg className="w-4 h-4 mr-1.5 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z" />
        </svg>
      ) : (
        <span className="mr-1.5">{icon}</span>
      )}
      {loading ? 'Processing…' : label}
    </Button>
  )
}

function AccountDetailSkeleton() {
  return (
    <div className="space-y-6 max-w-3xl">
      <Skeleton className="h-4 w-28" />
      <div className="rounded-xl border border-border bg-card p-6 space-y-3">
        <Skeleton className="h-6 w-40" />
        <Skeleton className="h-4 w-72" />
        <Skeleton className="h-10 w-24" />
      </div>
      <div className="flex gap-2">
        {[1, 2, 3, 4].map((i) => <Skeleton key={i} className="h-8 w-24" />)}
      </div>
    </div>
  )
}

// Icons
const DepositIcon = () => (
  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
  </svg>
)
const WithdrawIcon = () => (
  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M15 12H3m0 0 4-4m-4 4 4 4" />
  </svg>
)
const TransferIcon = () => (
  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M7.5 21 3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />
  </svg>
)
const ReconcileIcon = () => (
  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
  </svg>
)
const ListIcon = () => (
  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.8}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 6.75h12M8.25 12h12m-12 5.25h12M3.75 6.75h.007v.008H3.75V6.75Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0ZM3.75 12h.007v.008H3.75V12Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Zm-.375 5.25h.007v.008H3.75v-.008Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z" />
  </svg>
)
