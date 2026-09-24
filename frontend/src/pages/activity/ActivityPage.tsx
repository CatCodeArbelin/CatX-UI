import { useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router';
import dayjs, { type Dayjs } from 'dayjs';
import {
  Alert,
  Button,
  Card,
  DatePicker,
  Empty,
  Input,
  Space,
  Spin,
  Table,
  Tag,
  Tabs,
  Timeline,
  Typography,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { SearchOutlined } from '@ant-design/icons';
import { HttpUtil } from '@/utils';
import './ActivityPage.css';

interface ActivityEvent {
  id: number;
  observedAt: number;
  clientEmail?: string;
  domain?: string;
  destinationIp?: string;
  port?: number;
  protocol?: string;
  sni?: string;
  sessionKey?: string;
  source: string;
  provenance: string;
  confidence: number;
}

interface ActivitySession {
  id: number;
  sessionKey: string;
  clientEmail?: string;
  firstSeen: number;
  lastSeen: number;
  protocol?: string;
  source: string;
  provenance: string;
  confidence: number;
}

interface ActivityPageResponse<T> {
  enabled: boolean;
  items: T[];
  page: number;
  pageSize: number;
  total: number;
  from: number;
  to: number;
}

const DEFAULT_RANGE: [Dayjs, Dayjs] = [dayjs().subtract(7, 'day'), dayjs()];
const PAGE_SIZE = 25;
const TEXT = {
  title: 'Client activity',
  description: 'Observed destination and session metadata for one client.',
  clientEmail: 'Client email',
  enterClient: 'Enter a client email to view activity.',
  disabled: 'Analytics is disabled. Enable analytics to collect and view activity.',
  loadError: 'Analytics activity could not be loaded.',
  time: 'Time',
  destination: 'Destination',
  protocol: 'Protocol / port',
  source: 'Source / provenance',
  firstSeen: 'First seen',
  lastSeen: 'Last seen',
  events: 'Events',
  sessions: 'Sessions',
  range: 'Range',
  timeline: 'Timeline',
  sessionHistory: 'Session history',
  details: 'Details',
  noActivity: 'No observed activity in this range.',
  noSessions: 'No sessions in this range.',
};

function formatTime(value?: number): string {
  if (!value) return '—';
  return new Date(value).toLocaleString();
}

function provenanceTag(source: string, provenance: string) {
  return (
    <Tag>
      {source || 'unknown'} · {provenance || 'unknown'}
    </Tag>
  );
}

export default function ActivityPage() {
  const text = TEXT;
  const [searchParams, setSearchParams] = useSearchParams();
  const [emailInput, setEmailInput] = useState(searchParams.get('email') || '');
  const [email, setEmail] = useState(searchParams.get('email') || '');
  const [range, setRange] = useState<[Dayjs, Dayjs]>(DEFAULT_RANGE);
  const [activity, setActivity] = useState<ActivityPageResponse<ActivityEvent> | null>(null);
  const [sessions, setSessions] = useState<ActivityPageResponse<ActivitySession> | null>(null);
  const [loading, setLoading] = useState(Boolean(searchParams.get('email')));
  const [error, setError] = useState('');
  const [page, setPage] = useState(1);

  const params = useMemo(
    () => ({
      from: range[0].valueOf(),
      to: range[1].valueOf(),
      page,
      pageSize: PAGE_SIZE,
    }),
    [page, range],
  );

  useEffect(() => {
    if (!email) return;
    let cancelled = false;
    Promise.all([
      HttpUtil.get<ActivityPageResponse<ActivityEvent>>(
        `/panel/api/analytics/clients/${encodeURIComponent(email)}/activity`,
        params,
        { silent: true },
      ),
      HttpUtil.get<ActivityPageResponse<ActivitySession>>(
        `/panel/api/analytics/clients/${encodeURIComponent(email)}/sessions`,
        params,
        { silent: true },
      ),
    ])
      .then(([activityMsg, sessionMsg]) => {
        if (cancelled) return;
        if (!activityMsg.success || !sessionMsg.success) {
          setError(activityMsg.msg || sessionMsg.msg || TEXT.loadError);
          setActivity(null);
          setSessions(null);
          return;
        }
        setActivity(activityMsg.obj);
        setSessions(sessionMsg.obj);
      })
      .catch(() => {
        if (!cancelled) setError(TEXT.loadError);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [email, params]);

  const submit = () => {
    const next = emailInput.trim();
    setPage(1);
    setActivity(null);
    setSessions(null);
    setError('');
    setLoading(Boolean(next));
    setEmail(next);
    if (next) setSearchParams({ email: next });
    else setSearchParams({});
  };

  const eventColumns: ColumnsType<ActivityEvent> = [
    { title: text.time, dataIndex: 'observedAt', render: formatTime },
    { title: text.destination, render: (_, row) => row.domain || row.destinationIp || '—' },
    {
      title: text.protocol,
      render: (_, row) => (row.protocol ? `${row.protocol}${row.port ? `:${row.port}` : ''}` : '—'),
    },
    { title: text.source, render: (_, row) => provenanceTag(row.source, row.provenance) },
  ];
  const sessionColumns: ColumnsType<ActivitySession> = [
    { title: text.firstSeen, dataIndex: 'firstSeen', render: formatTime },
    { title: text.lastSeen, dataIndex: 'lastSeen', render: formatTime },
    { title: text.protocol, dataIndex: 'protocol', render: (value) => value || '—' },
    { title: text.source, render: (_, row) => provenanceTag(row.source, row.provenance) },
  ];

  const disabled = activity?.enabled === false || sessions?.enabled === false;
  const events = activity?.items || [];

  return (
    <div className="activity-page">
      <Card>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div>
            <Typography.Title level={2}>{text.title}</Typography.Title>
            <Typography.Paragraph type="secondary">{text.description}</Typography.Paragraph>
          </div>
          <Space wrap>
            <Input
              aria-label={text.clientEmail}
              placeholder={text.clientEmail}
              value={emailInput}
              onChange={(event) => setEmailInput(event.target.value)}
              onPressEnter={submit}
              style={{ width: 280 }}
            />
            <DatePicker.RangePicker
              showTime
              value={range}
              onChange={(next) => {
                if (next?.[0] && next?.[1]) {
                  setRange([next[0], next[1]]);
                  setPage(1);
                  if (email) setLoading(true);
                }
              }}
            />
            <Button type="primary" icon={<SearchOutlined />} onClick={submit}>
              Search
            </Button>
          </Space>
          {!email && <Empty description={text.enterClient} />}
          {disabled && <Alert type="info" showIcon message={text.disabled} />}
          {error && <Alert type="error" showIcon message={error} />}
          {email && !disabled && (
            <Spin spinning={loading}>
              <div className="activity-summary">
                <Tag>
                  {text.events}: {activity?.total || 0}
                </Tag>
                <Tag>
                  {text.sessions}: {sessions?.total || 0}
                </Tag>
                <Tag>
                  {text.range}: {formatTime(params.from)} – {formatTime(params.to)}
                </Tag>
              </div>
              <Tabs
                items={[
                  {
                    key: 'timeline',
                    label: text.timeline,
                    children: events.length ? (
                      <Timeline
                        items={events.map((event) => ({
                          key: event.id,
                          children: (
                            <div>
                              <Typography.Text strong>
                                {event.domain || event.destinationIp || '—'}
                              </Typography.Text>{' '}
                              · {event.protocol || '—'}
                              {event.port ? `:${event.port}` : ''}
                              <br />
                              <Typography.Text type="secondary">
                                {formatTime(event.observedAt)} · {event.source} · {event.provenance}
                              </Typography.Text>
                            </div>
                          ),
                        }))}
                      />
                    ) : (
                      <Empty description={text.noActivity} />
                    ),
                  },
                  {
                    key: 'sessions',
                    label: text.sessionHistory,
                    children: (
                      <Table<ActivitySession>
                        rowKey="sessionKey"
                        columns={sessionColumns}
                        dataSource={sessions?.items || []}
                        pagination={{
                          current: page,
                          pageSize: PAGE_SIZE,
                          total: sessions?.total || 0,
                          onChange: setPage,
                        }}
                        locale={{ emptyText: <Empty description={text.noSessions} /> }}
                      />
                    ),
                  },
                  {
                    key: 'details',
                    label: text.details,
                    children: (
                      <Table<ActivityEvent>
                        rowKey="id"
                        columns={eventColumns}
                        dataSource={events}
                        pagination={false}
                        locale={{ emptyText: <Empty description={text.noActivity} /> }}
                      />
                    ),
                  },
                ]}
              />
            </Spin>
          )}
        </Space>
      </Card>
    </div>
  );
}
