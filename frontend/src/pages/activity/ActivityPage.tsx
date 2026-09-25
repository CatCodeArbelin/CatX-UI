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
  InputNumber,
  Space,
  Spin,
  Switch,
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

type Page<T> = {
  enabled: boolean;
  items: T[];
  page: number;
  pageSize: number;
  total: number;
  from: number;
  to: number;
};
type Event = {
  id: number;
  observedAt: number;
  domain?: string;
  destinationIp?: string;
  service?: string;
  category?: string;
  asn?: number;
  country?: string;
  classificationSource?: string;
  classificationProvenance?: string;
  classificationConfidence?: number;
  classificationLevel?: string;
  classificationConflict?: boolean;
  classificationReason?: string;
  source: string;
  provenance: string;
  confidence: number;
};
type Session = {
  id: number;
  sessionKey: string;
  firstSeen: number;
  lastSeen: number;
  protocol?: string;
  source: string;
  provenance: string;
  confidence: number;
};
type DNS = {
  id: number;
  observedAt: number;
  domain: string;
  resolvedIp?: string;
  recordType?: string;
  source: string;
  provenance: string;
  confidence: number;
};
type Retention = {
  rawEvents: number;
  dnsObservations: number;
  sessions: number;
  aggregates: number;
};
type Settings = {
  enabled: boolean;
  dnsIntelligence: boolean;
  retention: Retention;
  privacy: { metadataOnly: boolean; classificationsMayBeUncertain: boolean };
};
type TrafficBreakdown = { name: string; observations: number; sessions: number };
type TrafficHistory = {
  from: number;
  to: number;
  up: number;
  down: number;
  clients: number;
  inbounds: number;
  nodes: number;
  serviceBreakdown: TrafficBreakdown[];
  categoryBreakdown: TrafficBreakdown[];
};

const RANGE: [Dayjs, Dayjs] = [dayjs().subtract(7, 'day'), dayjs()];
const SIZE = 25;
const text = {
  title: 'Client activity & DNS intelligence',
  description: 'Metadata-only observations for one existing panel client.',
  client: 'Client email',
  enter: 'Enter a client email to view activity.',
  disabled: 'Analytics is disabled. Enable analytics to collect and view activity.',
  dnsDisabled: 'DNS intelligence is disabled. Core analytics remains available.',
  error: 'Analytics activity could not be loaded.',
  privacy:
    'Privacy: metadata only. Classifications use visible metadata and may be uncertain; no payloads, cookies, credentials, or headers are collected.',
  noActivity: 'No observed activity in this range.',
  noDns: 'No DNS observations in this range.',
  noSessions: 'No sessions in this range.',
  noTraffic: 'No historical traffic in this range.',
};
const time = (value?: number) => (value ? new Date(value).toLocaleString() : '—');
const bytes = (value: number) => {
  if (!value) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const index = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1);
  return `${(value / 1024 ** index).toFixed(index ? 2 : 0)} ${units[index]}`;
};
const provenance = (source: string, value: string) => (
  <Tag>
    {source || 'unknown'} · {value || 'unknown'}
  </Tag>
);
const confidence = (level?: string, value?: number, conflict?: boolean) =>
  conflict ? (
    <Tag color="warning">Ambiguous / conflicting</Tag>
  ) : (
    <Tag>
      {level || 'unknown'} · {typeof value === 'number' ? `${Math.round(value * 100)}%` : '—'}
    </Tag>
  );

export default function ActivityPage() {
  const [query, setQuery] = useSearchParams();
  const [input, setInput] = useState(query.get('email') || '');
  const [email, setEmail] = useState(query.get('email') || '');
  const [range, setRange] = useState<[Dayjs, Dayjs]>(RANGE);
  const [events, setEvents] = useState<Page<Event> | null>(null);
  const [sessions, setSessions] = useState<Page<Session> | null>(null);
  const [dns, setDns] = useState<Page<DNS> | null>(null);
  const [settings, setSettings] = useState<Settings | null>(null);
  const [traffic, setTraffic] = useState<{ enabled: boolean; history: TrafficHistory } | null>(null);
  const [loading, setLoading] = useState(Boolean(email));
  const [error, setError] = useState('');
  const [settingsError, setSettingsError] = useState('');
  const [page, setPage] = useState(1);
  const params = useMemo(
    () => ({ from: range[0].valueOf(), to: range[1].valueOf(), page, pageSize: SIZE }),
    [page, range],
  );

  useEffect(() => {
    let cancelled = false;
    const settingsRequest = HttpUtil.get<Settings>('/panel/api/analytics/settings', undefined, {
      silent: true,
    });
    if (!email) {
      void settingsRequest.then((result) => {
        if (!cancelled && result.success && result.obj?.retention) setSettings(result.obj);
      });
      return () => {
        cancelled = true;
      };
    }
    void Promise.all([
      HttpUtil.get<Page<Event>>(
        `/panel/api/analytics/clients/${encodeURIComponent(email)}/activity`,
        params,
        { silent: true },
      ),
      HttpUtil.get<Page<Session>>(
        `/panel/api/analytics/clients/${encodeURIComponent(email)}/sessions`,
        params,
        { silent: true },
      ),
      HttpUtil.get<Page<DNS>>(
        `/panel/api/analytics/clients/${encodeURIComponent(email)}/dns`,
        params,
        { silent: true },
      ),
      HttpUtil.get<{ enabled: boolean; history: TrafficHistory }>(
        `/panel/api/analytics/clients/${encodeURIComponent(email)}/traffic`,
        { from: params.from, to: params.to },
        { silent: true },
      ),
      settingsRequest,
    ])
      .then(([eventResult, sessionResult, dnsResult, trafficResult, settingsResult]) => {
        if (cancelled) return;
        if (settingsResult.success && settingsResult.obj?.retention)
          setSettings(settingsResult.obj);
        if (!eventResult.success || !sessionResult.success || !dnsResult.success || !trafficResult.success) {
          setError(eventResult.msg || sessionResult.msg || dnsResult.msg || trafficResult.msg || text.error);
          return;
        }
        setEvents(eventResult.obj);
        setSessions(sessionResult.obj);
        setDns(dnsResult.obj);
        if (trafficResult.obj) setTraffic(trafficResult.obj);
      })
      .catch(() => {
        if (!cancelled) setError(text.error);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [email, params]);

  const submit = () => {
    const next = input.trim();
    setPage(1);
    setError('');
    setEvents(null);
    setSessions(null);
    setDns(null);
    setTraffic(null);
    setLoading(Boolean(next));
    setEmail(next);
    if (next) setQuery({ email: next });
    else setQuery({});
  };
  const saveSettings = async (value: Settings) => {
    setSettingsError('');
    const result = await HttpUtil.post<Settings>(
      '/panel/api/analytics/settings',
      { dnsIntelligence: value.dnsIntelligence, retention: value.retention },
      { silent: true },
    );
    if (result.success && result.obj) setSettings(result.obj);
    else setSettingsError(result.msg || 'Settings could not be saved.');
  };

  const eventColumns: ColumnsType<Event> = [
    { title: 'Time', dataIndex: 'observedAt', render: time },
    { title: 'Domain / IP', render: (_, row) => row.domain || row.destinationIp || '—' },
    { title: 'Service', render: (_, row) => row.service || '—' },
    { title: 'Category', render: (_, row) => row.category || 'unknown' },
    {
      title: 'ASN / country',
      render: (_, row) =>
        [row.asn ? `AS${row.asn}` : '', row.country].filter(Boolean).join(' · ') || '—',
    },
    {
      title: 'Source / provenance',
      render: (_, row) =>
        provenance(
          row.classificationSource || row.source,
          row.classificationProvenance || row.provenance,
        ),
    },
    {
      title: 'Confidence',
      render: (_, row) =>
        confidence(
          row.classificationLevel,
          row.classificationConfidence ?? row.confidence,
          row.classificationConflict,
        ),
    },
  ];
  const enrichmentFor = (row: DNS) =>
    (events?.items || []).find(
      (event) => event.domain === row.domain || event.destinationIp === row.resolvedIp,
    );
  const dnsColumns: ColumnsType<DNS> = [
    { title: 'Time', dataIndex: 'observedAt', render: time },
    { title: 'Observed domain', dataIndex: 'domain' },
    { title: 'Resolved IP', render: (_, row) => row.resolvedIp || '—' },
    { title: 'Service', render: (_, row) => enrichmentFor(row)?.service || '—' },
    { title: 'Category', render: (_, row) => enrichmentFor(row)?.category || 'unknown' },
    {
      title: 'ASN / country',
      render: (_, row) => {
        const event = enrichmentFor(row);
        return (
          [event?.asn ? `AS${event.asn}` : '', event?.country].filter(Boolean).join(' · ') || '—'
        );
      },
    },
    { title: 'Record', render: (_, row) => row.recordType || '—' },
    { title: 'Source / provenance', render: (_, row) => provenance(row.source, row.provenance) },
    { title: 'Confidence', render: (_, row) => confidence(undefined, row.confidence) },
  ];
  const sessionColumns: ColumnsType<Session> = [
    { title: 'First seen', dataIndex: 'firstSeen', render: time },
    { title: 'Last seen', dataIndex: 'lastSeen', render: time },
    { title: 'Protocol', dataIndex: 'protocol', render: (value) => value || '—' },
    { title: 'Source / provenance', render: (_, row) => provenance(row.source, row.provenance) },
  ];
  const analyticsDisabled = events?.enabled === false || sessions?.enabled === false;
  const dnsDisabled = dns?.enabled === false || settings?.dnsIntelligence === false;

  return (
    <div className="activity-page">
      <Card>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div>
            <Typography.Title level={2}>{text.title}</Typography.Title>
            <Typography.Paragraph type="secondary">{text.description}</Typography.Paragraph>
            <Alert type="info" showIcon message={text.privacy} />
          </div>
          <Space wrap>
            <Input
              aria-label={text.client}
              placeholder={text.client}
              value={input}
              onChange={(event) => setInput(event.target.value)}
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
          {!email && <Empty description={text.enter} />}
          {analyticsDisabled && <Alert type="info" showIcon message={text.disabled} />}
          {email && !analyticsDisabled && dnsDisabled && (
            <Alert type="warning" showIcon message={text.dnsDisabled} />
          )}
          {error && <Alert type="error" showIcon message={error} />}
          {email && !analyticsDisabled && (
            <Spin spinning={loading}>
              <div className="activity-summary">
                <Tag>Events: {events?.total || 0}</Tag>
                <Tag>DNS: {dns?.total || 0}</Tag>
                <Tag>Sessions: {sessions?.total || 0}</Tag>
                <Tag>
                  Range: {time(params.from)} – {time(params.to)}
                </Tag>
              </div>
              <Tabs
                items={[
                  {
                    key: 'timeline',
                    label: 'Timeline',
                    children: events?.items.length ? (
                      <Timeline
                        items={events.items.map((event) => ({
                          key: event.id,
                          children: (
                            <div>
                              <Typography.Text strong>
                                {event.domain || event.destinationIp || '—'}
                              </Typography.Text>{' '}
                              · {event.service || 'unknown'} · {event.category || 'unknown'}{' '}
                              {confidence(
                                event.classificationLevel,
                                event.classificationConfidence ?? event.confidence,
                                event.classificationConflict,
                              )}
                              <br />
                              <Typography.Text type="secondary">
                                {time(event.observedAt)} ·{' '}
                                {event.classificationReason || event.source}
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
                    key: 'dns',
                    label: 'DNS observations',
                    children: dnsDisabled ? (
                      <Empty description={text.dnsDisabled} />
                    ) : (
                      <Table<DNS>
                        rowKey="id"
                        columns={dnsColumns}
                        dataSource={dns?.items || []}
                        pagination={{
                          current: page,
                          pageSize: SIZE,
                          total: dns?.total || 0,
                          onChange: setPage,
                        }}
                        locale={{ emptyText: <Empty description={text.noDns} /> }}
                      />
                    ),
                  },
                  {
                    key: 'details',
                    label: 'Enriched details',
                    children: (
                      <Table<Event>
                        rowKey="id"
                        columns={eventColumns}
                        dataSource={events?.items || []}
                        pagination={false}
                        locale={{ emptyText: <Empty description={text.noActivity} /> }}
                      />
                    ),
                  },
                  {
                    key: 'sessions',
                    label: 'Session history',
                    children: (
                      <Table<Session>
                        rowKey="sessionKey"
                        columns={sessionColumns}
                        dataSource={sessions?.items || []}
                        pagination={{
                          current: page,
                          pageSize: SIZE,
                          total: sessions?.total || 0,
                          onChange: setPage,
                        }}
                        locale={{ emptyText: <Empty description={text.noSessions} /> }}
                      />
                    ),
                  },
                  {
                    key: 'traffic',
                    label: 'Traffic history',
                    children: traffic?.history ? (
                      <Space direction="vertical" style={{ width: '100%' }}>
                        <div className="activity-summary">
                          <Tag>Upload: {bytes(traffic.history.up)}</Tag>
                          <Tag>Download: {bytes(traffic.history.down)}</Tag>
                          <Tag>Clients: {traffic.history.clients}</Tag>
                          <Tag>Inbounds: {traffic.history.inbounds}</Tag>
                          <Tag>Nodes: {traffic.history.nodes}</Tag>
                        </div>
                        <Table<TrafficBreakdown>
                          rowKey="name"
                          columns={[
                            { title: 'Service / category', dataIndex: 'name' },
                            { title: 'Observations', dataIndex: 'observations' },
                            { title: 'Sessions', dataIndex: 'sessions' },
                          ]}
                          dataSource={[
                            ...traffic.history.serviceBreakdown,
                            ...traffic.history.categoryBreakdown,
                          ]}
                          pagination={false}
                          locale={{ emptyText: <Empty description={text.noTraffic} /> }}
                        />
                        <Typography.Text type="secondary">
                          Service and category rows are metadata-derived observations and sessions;
                          byte totals come only from upstream traffic accounting.
                        </Typography.Text>
                      </Space>
                    ) : (
                      <Empty description={text.noTraffic} />
                    ),
                  },
                ]}
              />
            </Spin>
          )}
          {settings && (
            <Card size="small" title="Privacy & retention">
              <Space direction="vertical" style={{ width: '100%' }}>
                <Typography.Text type="secondary">{text.privacy}</Typography.Text>
                <Space wrap>
                  <Typography.Text>Enable DNS intelligence</Typography.Text>
                  <Switch
                    checked={settings.dnsIntelligence}
                    disabled={!settings.enabled}
                    onChange={(checked) => setSettings({ ...settings, dnsIntelligence: checked })}
                  />
                  <Button onClick={() => void saveSettings(settings)}>Save settings</Button>
                </Space>
                <Space wrap>
                  <InputNumber
                    min={1}
                    max={3650}
                    addonBefore="Raw"
                    value={settings.retention.rawEvents}
                    onChange={(value) =>
                      setSettings({
                        ...settings,
                        retention: { ...settings.retention, rawEvents: value || 1 },
                      })
                    }
                  />
                  <InputNumber
                    min={1}
                    max={3650}
                    addonBefore="DNS"
                    value={settings.retention.dnsObservations}
                    onChange={(value) =>
                      setSettings({
                        ...settings,
                        retention: { ...settings.retention, dnsObservations: value || 1 },
                      })
                    }
                  />
                  <InputNumber
                    min={1}
                    max={3650}
                    addonBefore="Sessions"
                    value={settings.retention.sessions}
                    onChange={(value) =>
                      setSettings({
                        ...settings,
                        retention: { ...settings.retention, sessions: value || 1 },
                      })
                    }
                  />
                  <InputNumber
                    min={1}
                    max={3650}
                    addonBefore="Aggregates"
                    value={settings.retention.aggregates}
                    onChange={(value) =>
                      setSettings({
                        ...settings,
                        retention: { ...settings.retention, aggregates: value || 1 },
                      })
                    }
                  />
                </Space>
                {settingsError && <Alert type="error" message={settingsError} />}
              </Space>
            </Card>
          )}
        </Space>
      </Card>
    </div>
  );
}
