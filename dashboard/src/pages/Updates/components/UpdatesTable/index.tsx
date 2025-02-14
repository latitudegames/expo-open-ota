import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api.ts';
import { ApiError } from '@/components/APIError';
import { DataTable } from '@/components/DataTable';
import { GitBranch, Milestone, Rss } from 'lucide-react';
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

export const UpdatesTable = ({
  branch,
  runtimeVersion,
}: {
  branch: string;
  runtimeVersion: string;
}) => {
  const [, setSearchParams] = useSearchParams();
  const { data, isLoading, error } = useQuery({
    queryKey: ['updates'],
    queryFn: () => api.getUpdates(branch, runtimeVersion),
  });

  return (
    <div className="w-full flex-1">
      <Breadcrumb className="mb-2">
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink href="/" className="flex items-center gap-2 underline">
              <GitBranch className="w-4" />
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>{branch}</BreadcrumbPage>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbLink
              href={`/?branch=${branch}`}
              className="flex items-center gap-2 underline">
              <Milestone className="w-4" />
            </BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>{runtimeVersion}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      {!!error && <ApiError error={error} />}
      <DataTable
        loading={isLoading}
        columns={[
          {
            header: 'ID',
            accessorKey: 'updateId',
            cell: value => {
              return (
                <span className="flex flex-row gap-2 items-center w-full">
                  <Rss className="w-4" />
                  {value.row.original.updateId}
                </span>
              );
            },
          },
          {
            header: 'UUID',
            accessorKey: 'updateUUID',
            cell: value => {
              return value.row.original.updateUUID;
            },
          },
          {
            header: 'Platform',
            accessorKey: 'platform',
            cell: value => {
              const isIos = value.row.original.platform === 'ios';
              const isAndroid = value.row.original.platform === 'android';
              return (
                <div className="flex flex-row items-center gap-2">
                  {isIos && <img src="/apple.svg" className="w-4" alt="apple" />}
                  {isAndroid && <img src="/android.svg" className="w-4" alt="android" />}
                </div>
              );
            },
          },
          {
            header: 'Commit',
            accessorKey: 'commitHash',
            cell: value => {
              return (
                <Badge variant="secondary" className="text-xs">
                  {value.row.original.commitHash.slice(0, 7)}
                </Badge>
              );
            },
          },
          {
            header: 'Published at',
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
        ]}
        data={data ?? []}
      />
    </div>
  );
};
