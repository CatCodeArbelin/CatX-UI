// CatX route adapter: established sponsor presentation stays reusable while
// route ownership and data access remain in the fork registry/query.
import SponsorsPage from '@/pages/sponsors/SponsorsPage';

export default function CatxSponsorsPage() {
  return <SponsorsPage showManagementAction />;
}
