import { useQuery } from '@tanstack/react-query';

import { HttpUtil } from '@/utils';
import { keys } from '@/api/queryKeys';
import type { SponsorList } from '@/generated/types';

const EMPTY: SponsorList = { sponsors: [] };

async function fetchSponsors(): Promise<SponsorList> {
  const msg = await HttpUtil.get<SponsorList>('/panel/api/fork/sponsors', undefined, {
    silent: true,
  });
  if (!msg?.success || !msg.obj) throw new Error(msg?.msg || 'Failed to load sponsors');
  return { contact: msg.obj.contact, sponsors: msg.obj.sponsors ?? [] };
}

export function useSponsorsQuery() {
  const query = useQuery({
    queryKey: keys.sponsors(),
    queryFn: fetchSponsors,
    staleTime: 60 * 60 * 1000,
    retry: false,
  });
  return {
    data: query.data ?? EMPTY,
    fetched: query.isFetched,
    error:
      query.error instanceof Error
        ? query.error.message
        : query.error
          ? 'Failed to load sponsors'
          : undefined,
  };
}
