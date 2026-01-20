import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { api } from '@/lib/api.ts';
import { ApiError } from '@/components/APIError';
import { DataTable } from '@/components/DataTable';
import { GitBranch, Milestone, Trash2 } from 'lucide-react';
import { useSearchParams } from 'react-router';
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from '@/components/ui/breadcrumb';
import { Badge } from '@/components/ui/badge.tsx';
import { Button } from '@/components/ui/button.tsx';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { toast } from '@/hooks/use-toast';

export const RuntimeVersionsTable = ({ branch }: { branch: string }) => {
  const [, setSearchParams] = useSearchParams();
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedRuntimeVersion, setSelectedRuntimeVersion] = useState<string | null>(null);
  const [selectedUpdateCount, setSelectedUpdateCount] = useState<number>(0);
  const queryClient = useQueryClient();

  const { data, isLoading, error } = useQuery({
    queryKey: ['runtimeVersions', branch],
    queryFn: () => api.getRuntimeVersions(branch),
  });

  const deleteMutation = useMutation({
    mutationFn: (runtimeVersion: string) => api.deleteRuntimeVersion(branch, runtimeVersion),
    onSuccess: (result) => {
      toast({
        title: 'Runtime version deleted',
        description: `Successfully deleted ${result.deletedCount} of ${result.totalCount} updates.`,
      });
      queryClient.invalidateQueries({ queryKey: ['runtimeVersions', branch] });
      setDeleteDialogOpen(false);
      setSelectedRuntimeVersion(null);
    },
    onError: (error) => {
      toast({
        title: 'Delete failed',
        description: error instanceof Error ? error.message : 'Failed to delete runtime version',
        variant: 'destructive',
      });
    },
  });

  const handleDeleteClick = (runtimeVersion: string, updateCount: number) => {
    setSelectedRuntimeVersion(runtimeVersion);
    setSelectedUpdateCount(updateCount);
    setDeleteDialogOpen(true);
  };

  const confirmDelete = () => {
    if (selectedRuntimeVersion) {
      deleteMutation.mutate(selectedRuntimeVersion);
    }
  };

  return (
    <div className="w-full flex-1">
      <Breadcrumb className="mb-2">
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink href="/dashboard" className="flex items-center gap-2">
              <GitBranch className="w-4" />
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>{branch}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      {!!error && <ApiError error={error} />}
      <DataTable
        loading={isLoading}
        columns={[
          {
            header: 'Runtime version',
            accessorKey: 'runtimeVersion',
            cell: value => {
              return (
                <button
                  className="flex flex-row gap-2 items-center cursor-pointer w-full underline"
                  onClick={() => {
                    setSearchParams({
                      branch,
                      runtimeVersion: value.row.original.runtimeVersion,
                    });
                  }}>
                  <Milestone className="w-4" />
                  {value.row.original.runtimeVersion}
                </button>
              );
            },
          },
          {
            header: 'Created at',
            accessorKey: 'createdAt',
            cell: ({ row }) => {
              const date = new Date(row.original.createdAt);
              return (
                <Badge variant="outline">
                  {date.toLocaleDateString('en-GB', {
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric',
                    hour: 'numeric',
                    minute: 'numeric',
                    second: 'numeric',
                  })}
                </Badge>
              );
            },
          },
          {
            header: 'Last update',
            accessorKey: 'lastUpdatedAt',
            cell: ({ row }) => {
              const date = new Date(row.original.lastUpdatedAt);
              return (
                <Badge variant="outline">
                  {date.toLocaleDateString('en-GB', {
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric',
                    hour: 'numeric',
                    minute: 'numeric',
                    second: 'numeric',
                  })}
                </Badge>
              );
            },
          },
          {
            header: '# Updates',
            accessorKey: 'numberOfUpdates',
            cell: ({ row }) => {
              return <Badge variant="secondary">{row.original.numberOfUpdates}</Badge>;
            },
          },
          {
            header: '',
            accessorKey: 'actions',
            cell: ({ row }) => {
              return (
                <Button
                  variant="ghost"
                  size="icon"
                  className="text-destructive hover:text-destructive hover:bg-destructive/10"
                  onClick={(e) => {
                    e.stopPropagation();
                    handleDeleteClick(row.original.runtimeVersion, row.original.numberOfUpdates);
                  }}
                  title="Delete all updates for this runtime version"
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              );
            },
          },
        ]}
        data={data ?? []}
      />

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Runtime Version</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete all {selectedUpdateCount} update{selectedUpdateCount !== 1 ? 's' : ''} for
              runtime version <strong>{selectedRuntimeVersion}</strong>? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDeleteDialogOpen(false)}
              disabled={deleteMutation.isPending}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={confirmDelete}
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
};
