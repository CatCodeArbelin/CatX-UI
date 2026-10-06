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
import ForkAdminPageShell from '@/components/fork/ForkAdminPageShell';
import CatxState from '@/components/fork/CatxState';
import { i18n } from '@/i18n/react';
import { useTranslation } from 'react-i18next';
import type { ForkRuntimeState } from '@/lib/fork-feature';
import './ActivityPage.css';

type Page<T> = {
  enabled: boolean;
  state?: ForkRuntimeState;
  featureDisabled?: boolean;
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
  state?: ForkRuntimeState;
  dnsState?: ForkRuntimeState;
  featureDisabled?: boolean;
  restartRequired?: boolean;
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
const time = (value?: number) => (value ? new Date(value).toLocaleString() : '—');
const bytes = (value: number) => {
  if (!value) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const index = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1);
  return `${(value / 1024 ** index).toFixed(index ? 2 : 0)} ${units[index]}`;
};
const provenance = (source: string, value: string) => (
  <Tag>
    {source || i18n.t('fork.common.unknown')} · {value || i18n.t('fork.common.unknown')}
  </Tag>
);
const confidence = (level?: string, value?: number, conflict?: boolean) =>
  conflict ? (
    <Tag color="warning">{i18n.t('fork.activity.labels.ambiguous')}</Tag>
  ) : (
    <Tag>
      {level || i18n.t('fork.common.unknown')} ·{' '}
      {typeof value === 'number' ? `${Math.round(value * 100)}%` : '—'}
    </Tag>
  );

export default function ActivityPage() {
  const { t } = useTranslation();
  const [query, setQuery] = useSearchParams();
  const [input, setInput] = useState(query.get('email') || '');
  const [email, setEmail] = useState(query.get('email') || '');
  const [range, setRange] = useState<[Dayjs, Dayjs]>(RANGE);
  const [events, setEvents] = useState<Page<Event> | null>(null);
  const [sessions, setSessions] = useState<Page<Session> | null>(null);
  const [dns, setDns] = useState<Page<DNS> | null>(null);
  const [settings, setSettings] = useState<Settings | null>(null);
  const [traffic, setTraffic] = useState<{ enabled: boolean; history: TrafficHistory } | null>(
    null,
  );
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
        if (cancelled) return;
        if (result.success && result.obj?.retention) setSettings(result.obj);
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
        if (
          !eventResult.success ||
          !sessionResult.success ||
          !dnsResult.success ||
          !trafficResult.success
        ) {
          setEvents(null);
          setSessions(null);
          setDns(null);
          setTraffic(null);
          setError(t('fork.activity.error'));
          return;
        }
        setEvents(eventResult.obj);
        setSessions(sessionResult.obj);
        setDns(dnsResult.obj);
        if (trafficResult.obj) setTraffic(trafficResult.obj);
      })
      .catch(() => {
        if (!cancelled) {
          setEvents(null);
          setSessions(null);
          setDns(null);
          setTraffic(null);
          setError(t('fork.activity.error'));
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [email, params, t]);

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
      { headers: { 'Content-Type': 'application/json' }, silent: true },
    );
    if (result.success && result.obj) setSettings(result.obj);
    else setSettingsError(t('fork.activity.labels.saveError'));
  };

  const eventColumns: ColumnsType<Event> = [
    { title: t('fork.activity.labels.time'), dataIndex: 'observedAt', render: time },
    {
      title: t('fork.activity.labels.domainIp'),
      render: (_, row) => row.domain || row.destinationIp || '—',
    },
    { title: t('fork.activity.labels.service'), render: (_, row) => row.service || '—' },
    {
      title: t('fork.activity.labels.category'),
      render: (_, row) => row.category || t('fork.common.unknown'),
    },
    {
      title: t('fork.activity.labels.asnCountry'),
      render: (_, row) =>
        [row.asn ? `AS${row.asn}` : '', row.country].filter(Boolean).join(' · ') || '—',
    },
    {
      title: t('fork.activity.labels.sourceProvenance'),
      render: (_, row) =>
        provenance(
          row.classificationSource || row.source,
          row.classificationProvenance || row.provenance,
        ),
    },
    {
      title: t('fork.activity.labels.confidence'),
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
    { title: t('fork.activity.labels.time'), dataIndex: 'observedAt', render: time },
    { title: t('fork.activity.labels.observedDomain'), dataIndex: 'domain' },
    { title: t('fork.activity.labels.resolvedIp'), render: (_, row) => row.resolvedIp || '—' },
    {
      title: t('fork.activity.labels.service'),
      render: (_, row) => enrichmentFor(row)?.service || '—',
    },
    {
      title: t('fork.activity.labels.category'),
      render: (_, row) => enrichmentFor(row)?.category || t('fork.common.unknown'),
    },
    {
      title: t('fork.activity.labels.asnCountry'),
      render: (_, row) => {
        const event = enrichmentFor(row);
        return (
          [event?.asn ? `AS${event.asn}` : '', event?.country].filter(Boolean).join(' · ') || '—'
        );
      },
    },
    { title: t('fork.activity.labels.record'), render: (_, row) => row.recordType || '—' },
    {
      title: t('fork.activity.labels.sourceProvenance'),
      render: (_, row) => provenance(row.source, row.provenance),
    },
    {
      title: t('fork.activity.labels.confidence'),
      render: (_, row) => confidence(undefined, row.confidence),
    },
  ];
  const sessionColumns: ColumnsType<Session> = [
    { title: t('fork.activity.labels.firstSeen'), dataIndex: 'firstSeen', render: time },
    { title: t('fork.activity.labels.lastSeen'), dataIndex: 'lastSeen', render: time },
    {
      title: t('fork.activity.labels.protocol'),
      dataIndex: 'protocol',
      render: (value) => value || '—',
    },
    {
      title: t('fork.activity.labels.sourceProvenance'),
      render: (_, row) => provenance(row.source, row.provenance),
    },
  ];
  const analyticsState: ForkRuntimeState =
    settings?.state ||
    events?.state ||
    sessions?.state ||
    (settings?.enabled === false || events?.enabled === false || sessions?.enabled === false
      ? 'feature_off'
      : 'active');
  const dnsState: ForkRuntimeState =
    settings?.dnsState ||
    dns?.state ||
    (dns?.enabled === false || settings?.dnsIntelligence === false ? 'feature_off' : 'active');
  const analyticsDisabled = analyticsState !== 'active';
  const dnsDisabled = dnsState !== 'active';
  const analyticsRequestFailed = Boolean(email && error);

  return (
    <ForkAdminPageShell pageClass="activity-page">
      <Card size="small">
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div>
            <Typography.Title level={2}>{t('fork.activity.title')}</Typography.Title>
            <Typography.Paragraph type="secondary">
              {t('fork.activity.subtitle')}
            </Typography.Paragraph>
            <Alert type="info" showIcon message={t('fork.activity.privacy')} />
          </div>
          {!analyticsDisabled && (
            <>
              <Space wrap>
                <Input
                  aria-label={t('fork.activity.client')}
                  placeholder={t('fork.activity.client')}
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
                  {t('fork.common.labels.search')}
                </Button>
              </Space>
              {!email && <Empty description={t('fork.activity.enter')} />}
            </>
          )}
          {analyticsDisabled && <CatxState state={analyticsState} feature="analytics" />}
          {email && !analyticsDisabled && dnsDisabled && (
            <CatxState state={dnsState} feature="dns_intelligence" />
          )}
          {analyticsRequestFailed ? (
            <CatxState state="error" feature="analytics" description={error} />
          ) : null}
          {email && !analyticsDisabled && !analyticsRequestFailed && (
            <Spin spinning={loading}>
              <div className="activity-summary">
                <Tag>
                  {t('fork.activity.labels.events')}: {events?.total || 0}
                </Tag>
                <Tag>
                  {t('fork.activity.labels.dns')}: {dns?.total || 0}
                </Tag>
                <Tag>
                  {t('fork.activity.labels.sessions')}: {sessions?.total || 0}
                </Tag>
                <Tag>
                  {t('fork.activity.labels.range')}: {time(params.from)} – {time(params.to)}
                </Tag>
              </div>
              <Tabs
                items={[
                  {
                    key: 'timeline',
                    label: t('fork.activity.labels.timeline'),
                    children: events?.items.length ? (
                      <Timeline
                        items={events.items.map((event) => ({
                          key: event.id,
                          children: (
                            <div>
                              <Typography.Text strong>
                                {event.domain || event.destinationIp || '—'}
                              </Typography.Text>{' '}
                              · {event.service || t('fork.common.unknown')} ·{' '}
                              {event.category || t('fork.common.unknown')}{' '}
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
                      <Empty description={t('fork.activity.noActivity')} />
                    ),
                  },
                  {
                    key: 'dns',
                    label: t('fork.activity.labels.dnsObservations'),
                    children: dnsDisabled ? (
                      <Empty description={t('fork.activity.dnsDisabled')} />
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
                        locale={{ emptyText: <Empty description={t('fork.activity.noDns')} /> }}
                      />
                    ),
                  },
                  {
                    key: 'details',
                    label: t('fork.activity.labels.enrichedDetails'),
                    children: (
                      <Table<Event>
                        rowKey="id"
                        columns={eventColumns}
                        dataSource={events?.items || []}
                        pagination={false}
                        locale={{
                          emptyText: <Empty description={t('fork.activity.noActivity')} />,
                        }}
                      />
                    ),
                  },
                  {
                    key: 'sessions',
                    label: t('fork.activity.labels.sessionHistory'),
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
                        locale={{
                          emptyText: <Empty description={t('fork.activity.noSessions')} />,
                        }}
                      />
                    ),
                  },
                  {
                    key: 'traffic',
                    label: t('fork.activity.labels.trafficHistory'),
                    children: traffic?.history ? (
                      <Space direction="vertical" style={{ width: '100%' }}>
                        <div className="activity-summary">
                          <Tag>
                            {t('fork.activity.labels.upload')}: {bytes(traffic.history.up)}
                          </Tag>
                          <Tag>
                            {t('fork.activity.labels.download')}: {bytes(traffic.history.down)}
                          </Tag>
                          <Tag>
                            {t('fork.activity.labels.clients')}: {traffic.history.clients}
                          </Tag>
                          <Tag>
                            {t('fork.activity.labels.inbounds')}: {traffic.history.inbounds}
                          </Tag>
                          <Tag>
                            {t('fork.activity.labels.nodes')}: {traffic.history.nodes}
                          </Tag>
                        </div>
                        <Table<TrafficBreakdown>
                          rowKey="name"
                          columns={[
                            { title: t('fork.activity.labels.serviceCategory'), dataIndex: 'name' },
                            {
                              title: t('fork.activity.labels.observations'),
                              dataIndex: 'observations',
                            },
                            { title: t('fork.activity.labels.sessions'), dataIndex: 'sessions' },
                          ]}
                          dataSource={[
                            ...traffic.history.serviceBreakdown,
                            ...traffic.history.categoryBreakdown,
                          ]}
                          pagination={false}
                          locale={{
                            emptyText: <Empty description={t('fork.activity.noTraffic')} />,
                          }}
                        />
                        <Typography.Text type="secondary">
                          {t('fork.activity.labels.metadataNote')}
                        </Typography.Text>
                      </Space>
                    ) : (
                      <Empty description={t('fork.activity.noTraffic')} />
                    ),
                  },
                ]}
              />
            </Spin>
          )}
          {settings && !analyticsDisabled && (
            <Card size="small" title={t('fork.activity.retention')}>
              <Space direction="vertical" style={{ width: '100%' }}>
                <Typography.Text type="secondary">{t('fork.activity.privacy')}</Typography.Text>
                <Space wrap>
                  <Typography.Text>{t('fork.common.labels.enableDns')}</Typography.Text>
                  <Switch
                    checked={settings.dnsIntelligence}
                    disabled={!settings.enabled}
                    onChange={(checked) => setSettings({ ...settings, dnsIntelligence: checked })}
                  />
                  <Button onClick={() => void saveSettings(settings)}>
                    {t('fork.common.labels.save')}
                  </Button>
                </Space>
                <Space wrap>
                  <InputNumber
                    min={1}
                    max={3650}
                    addonBefore={t('fork.common.labels.raw')}
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
                    addonBefore={t('fork.activity.labels.sessions')}
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
                    addonBefore={t('fork.activity.labels.aggregates')}
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
    </ForkAdminPageShell>
  );
}
