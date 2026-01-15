import * as duckdb from '@duckdb/duckdb-wasm';
import duckdb_wasm from '@duckdb/duckdb-wasm/dist/duckdb-mvp.wasm?url';
import mvp_worker from '@duckdb/duckdb-wasm/dist/duckdb-browser-mvp.worker.js?url';
import duckdb_wasm_eh from '@duckdb/duckdb-wasm/dist/duckdb-eh.wasm?url';
import eh_worker from '@duckdb/duckdb-wasm/dist/duckdb-browser-eh.worker.js?url';

let db: duckdb.AsyncDuckDB | null = null;
let conn: duckdb.AsyncDuckDBConnection | null = null;

const MANUAL_BUNDLES: duckdb.DuckDBBundles = {
	mvp: {
		mainModule: duckdb_wasm,
		mainWorker: mvp_worker
	},
	eh: {
		mainModule: duckdb_wasm_eh,
		mainWorker: eh_worker
	}
};

export async function initDuckDB(): Promise<duckdb.AsyncDuckDB> {
	if (db) return db;

	const bundle = await duckdb.selectBundle(MANUAL_BUNDLES);
	const worker = new Worker(bundle.mainWorker!);
	const logger = new duckdb.ConsoleLogger();
	db = new duckdb.AsyncDuckDB(logger, worker);
	await db.instantiate(bundle.mainModule, bundle.pthreadWorker);

	return db;
}

export async function getConnection(): Promise<duckdb.AsyncDuckDBConnection> {
	if (conn) return conn;

	const database = await initDuckDB();
	conn = await database.connect();
	return conn;
}

export async function query<T = Record<string, unknown>>(sql: string): Promise<T[]> {
	const connection = await getConnection();
	const result = await connection.query(sql);
	return result.toArray().map((row) => Object.fromEntries(row) as T);
}

export async function loadParquetFiles(): Promise<void> {
	const database = await initDuckDB();

	// Register parquet files from static folder
	await database.registerFileURL(
		'sessions.parquet',
		'/data/sessions.parquet',
		duckdb.DuckDBDataProtocol.HTTP,
		false
	);
	await database.registerFileURL(
		'messages.parquet',
		'/data/messages.parquet',
		duckdb.DuckDBDataProtocol.HTTP,
		false
	);
	await database.registerFileURL(
		'tool_calls.parquet',
		'/data/tool_calls.parquet',
		duckdb.DuckDBDataProtocol.HTTP,
		false
	);

	// Create views for easier querying
	const connection = await getConnection();
	await connection.query(`CREATE VIEW IF NOT EXISTS sessions AS SELECT * FROM 'sessions.parquet'`);
	await connection.query(`CREATE VIEW IF NOT EXISTS messages AS SELECT * FROM 'messages.parquet'`);
	await connection.query(
		`CREATE VIEW IF NOT EXISTS tool_calls AS SELECT * FROM 'tool_calls.parquet'`
	);
}

export type HeatmapData = {
	day_of_week: number;
	hour: number;
	count: number;
};

export async function getActivityHeatmap(): Promise<HeatmapData[]> {
	const connection = await getConnection();
	const result = await connection.query(`
		SELECT
			CAST(EXTRACT(DOW FROM timestamp) AS INTEGER) as day_of_week,
			CAST(EXTRACT(HOUR FROM timestamp) AS INTEGER) as hour,
			CAST(COUNT(*) AS INTEGER) as count
		FROM messages
		GROUP BY day_of_week, hour
		ORDER BY day_of_week, hour
	`);
	return result.toArray().map((row) => ({
		day_of_week: Number(row.day_of_week),
		hour: Number(row.hour),
		count: Number(row.count)
	}));
}

export type DailyCount = {
	date: string;
	count: number;
};

export async function getMessageTrend(): Promise<DailyCount[]> {
	const connection = await getConnection();
	const result = await connection.query(`
		SELECT
			CAST(DATE_TRUNC('day', timestamp) AS DATE) as date,
			CAST(COUNT(*) AS INTEGER) as count
		FROM messages
		GROUP BY date
		ORDER BY date
	`);
	return result.toArray().map((row) => ({
		date: String(row.date),
		count: Number(row.count)
	}));
}

export type ToolCount = {
	tool_name: string;
	count: number;
};

export async function getTopTools(limit: number = 15): Promise<ToolCount[]> {
	const connection = await getConnection();
	const result = await connection.query(`
		SELECT
			tool_name,
			CAST(COUNT(*) AS INTEGER) as count
		FROM tool_calls
		GROUP BY tool_name
		ORDER BY count DESC
		LIMIT ${limit}
	`);
	return result.toArray().map((row) => ({
		tool_name: String(row.tool_name),
		count: Number(row.count)
	}));
}

export async function getStats(): Promise<{
	sessions: number;
	messages: number;
	toolCalls: number;
}> {
	const connection = await getConnection();

	const sessionsResult = await connection.query('SELECT COUNT(*) as count FROM sessions');
	const messagesResult = await connection.query('SELECT COUNT(*) as count FROM messages');
	const toolCallsResult = await connection.query('SELECT COUNT(*) as count FROM tool_calls');

	return {
		sessions: Number(sessionsResult.toArray()[0].count),
		messages: Number(messagesResult.toArray()[0].count),
		toolCalls: Number(toolCallsResult.toArray()[0].count)
	};
}

// ============================================
// Additional Query Functions for Dashboard
// ============================================

export type RoleCount = {
	role: string;
	count: number;
};

export async function getMessagesByRole(): Promise<RoleCount[]> {
	const connection = await getConnection();
	// Note: messages table uses 'type' column, not 'role'
	const result = await connection.query(`
		SELECT
			type as role,
			CAST(COUNT(*) AS INTEGER) as count
		FROM messages
		WHERE type IN ('user', 'assistant')
		GROUP BY type
		ORDER BY count DESC
	`);
	return result.toArray().map((row) => ({
		role: String(row.role),
		count: Number(row.count)
	}));
}

export type LengthBucket = {
	bucket: string;
	count: number;
};

export async function getMessageLengthDistribution(): Promise<LengthBucket[]> {
	const connection = await getConnection();
	const result = await connection.query(`
		SELECT
			CASE
				WHEN LENGTH(content) < 100 THEN '< 100'
				WHEN LENGTH(content) < 500 THEN '100-500'
				WHEN LENGTH(content) < 1000 THEN '500-1K'
				WHEN LENGTH(content) < 5000 THEN '1K-5K'
				WHEN LENGTH(content) < 10000 THEN '5K-10K'
				ELSE '10K+'
			END as bucket,
			CAST(COUNT(*) AS INTEGER) as count
		FROM messages
		GROUP BY bucket
		ORDER BY
			CASE bucket
				WHEN '< 100' THEN 1
				WHEN '100-500' THEN 2
				WHEN '500-1K' THEN 3
				WHEN '1K-5K' THEN 4
				WHEN '5K-10K' THEN 5
				ELSE 6
			END
	`);
	return result.toArray().map((row) => ({
		bucket: String(row.bucket),
		count: Number(row.count)
	}));
}

export type SessionSummary = {
	session_id: string;
	project_name: string;
	start_time: string;
	end_time: string;
	message_count: number;
	duration_minutes: number;
};

export async function getSessionList(limit: number = 20, offset: number = 0): Promise<SessionSummary[]> {
	const connection = await getConnection();
	// Use sessions table which has pre-aggregated data
	const result = await connection.query(`
		SELECT
			session_id,
			project_name,
			first_message_at as start_time,
			last_message_at as end_time,
			CAST(message_count AS INTEGER) as message_count,
			CAST(EXTRACT(EPOCH FROM (last_message_at - first_message_at)) / 60 AS INTEGER) as duration_minutes
		FROM sessions
		ORDER BY first_message_at DESC
		LIMIT ${limit} OFFSET ${offset}
	`);
	return result.toArray().map((row) => ({
		session_id: String(row.session_id),
		project_name: String(row.project_name),
		start_time: String(row.start_time),
		end_time: String(row.end_time),
		message_count: Number(row.message_count),
		duration_minutes: Number(row.duration_minutes)
	}));
}

export type SessionDurationStats = {
	avg_duration_minutes: number;
	max_duration_minutes: number;
	total_sessions: number;
	total_hours: number;
};

export async function getSessionDurationStats(): Promise<SessionDurationStats> {
	const connection = await getConnection();
	// Use sessions table which has pre-computed timestamps
	const result = await connection.query(`
		WITH session_durations AS (
			SELECT
				session_id,
				EXTRACT(EPOCH FROM (last_message_at - first_message_at)) / 60 as duration_minutes
			FROM sessions
		)
		SELECT
			CAST(AVG(duration_minutes) AS INTEGER) as avg_duration_minutes,
			CAST(MAX(duration_minutes) AS INTEGER) as max_duration_minutes,
			CAST(COUNT(*) AS INTEGER) as total_sessions,
			CAST(SUM(duration_minutes) / 60 AS INTEGER) as total_hours
		FROM session_durations
	`);
	const row = result.toArray()[0];
	return {
		avg_duration_minutes: Number(row.avg_duration_minutes),
		max_duration_minutes: Number(row.max_duration_minutes),
		total_sessions: Number(row.total_sessions),
		total_hours: Number(row.total_hours)
	};
}

export type ToolTrend = {
	date: string;
	count: number;
};

export async function getToolTrend(): Promise<ToolTrend[]> {
	const connection = await getConnection();
	const result = await connection.query(`
		SELECT
			CAST(DATE_TRUNC('day', timestamp) AS DATE) as date,
			CAST(COUNT(*) AS INTEGER) as count
		FROM tool_calls
		GROUP BY date
		ORDER BY date
	`);
	return result.toArray().map((row) => ({
		date: String(row.date),
		count: Number(row.count)
	}));
}

export type ProjectCount = {
	project_name: string;
	message_count: number;
	session_count: number;
};

export async function getProjectBreakdown(): Promise<ProjectCount[]> {
	const connection = await getConnection();
	// Use sessions table which has project_name and message_count
	const result = await connection.query(`
		SELECT
			project_name,
			CAST(SUM(message_count) AS INTEGER) as message_count,
			CAST(COUNT(*) AS INTEGER) as session_count
		FROM sessions
		GROUP BY project_name
		ORDER BY message_count DESC
		LIMIT 10
	`);
	return result.toArray().map((row) => ({
		project_name: String(row.project_name),
		message_count: Number(row.message_count),
		session_count: Number(row.session_count)
	}));
}

// ============================================
// Context Window Analysis Queries
// ============================================

export type TokenEstimate = {
	session_id: string;
	total_chars: number;
	estimated_tokens: number;
	message_count: number;
};

export async function getTokenEstimatesPerSession(limit: number = 20): Promise<TokenEstimate[]> {
	const connection = await getConnection();
	// Rough token estimate: ~4 chars per token for English text
	const result = await connection.query(`
		SELECT
			session_id,
			CAST(SUM(LENGTH(content)) AS INTEGER) as total_chars,
			CAST(SUM(LENGTH(content)) / 4 AS INTEGER) as estimated_tokens,
			CAST(COUNT(*) AS INTEGER) as message_count
		FROM messages
		GROUP BY session_id
		ORDER BY total_chars DESC
		LIMIT ${limit}
	`);
	return result.toArray().map((row) => ({
		session_id: String(row.session_id),
		total_chars: Number(row.total_chars),
		estimated_tokens: Number(row.estimated_tokens),
		message_count: Number(row.message_count)
	}));
}

export type ContentTypeCount = {
	content_type: string;
	count: number;
};

export async function getContentTypeDistribution(): Promise<ContentTypeCount[]> {
	const connection = await getConnection();
	const result = await connection.query(`
		SELECT
			COALESCE(content_type, 'unknown') as content_type,
			CAST(COUNT(*) AS INTEGER) as count
		FROM messages
		GROUP BY content_type
		ORDER BY count DESC
	`);
	return result.toArray().map((row) => ({
		content_type: String(row.content_type),
		count: Number(row.count)
	}));
}

export type LargeMessageInfo = {
	session_id: string;
	type: string;
	char_count: number;
	estimated_tokens: number;
	preview: string;
};

export async function getLargeMessages(minChars: number = 10000, limit: number = 20): Promise<LargeMessageInfo[]> {
	const connection = await getConnection();
	const result = await connection.query(`
		SELECT
			session_id,
			type,
			CAST(LENGTH(content) AS INTEGER) as char_count,
			CAST(LENGTH(content) / 4 AS INTEGER) as estimated_tokens,
			LEFT(content, 100) as preview
		FROM messages
		WHERE LENGTH(content) > ${minChars}
		ORDER BY char_count DESC
		LIMIT ${limit}
	`);
	return result.toArray().map((row) => ({
		session_id: String(row.session_id),
		type: String(row.type),
		char_count: Number(row.char_count),
		estimated_tokens: Number(row.estimated_tokens),
		preview: String(row.preview)
	}));
}

export type ContextStats = {
	total_chars: number;
	total_estimated_tokens: number;
	avg_chars_per_session: number;
	avg_tokens_per_session: number;
	largest_session_tokens: number;
	sessions_over_100k_tokens: number;
};

export async function getContextStats(): Promise<ContextStats> {
	const connection = await getConnection();
	const result = await connection.query(`
		WITH session_sizes AS (
			SELECT
				session_id,
				SUM(LENGTH(content)) as total_chars,
				SUM(LENGTH(content)) / 4 as estimated_tokens
			FROM messages
			GROUP BY session_id
		)
		SELECT
			CAST(SUM(total_chars) AS BIGINT) as total_chars,
			CAST(SUM(estimated_tokens) AS BIGINT) as total_estimated_tokens,
			CAST(AVG(total_chars) AS INTEGER) as avg_chars_per_session,
			CAST(AVG(estimated_tokens) AS INTEGER) as avg_tokens_per_session,
			CAST(MAX(estimated_tokens) AS INTEGER) as largest_session_tokens,
			CAST(SUM(CASE WHEN estimated_tokens > 100000 THEN 1 ELSE 0 END) AS INTEGER) as sessions_over_100k_tokens
		FROM session_sizes
	`);
	const row = result.toArray()[0];
	return {
		total_chars: Number(row.total_chars),
		total_estimated_tokens: Number(row.total_estimated_tokens),
		avg_chars_per_session: Number(row.avg_chars_per_session),
		avg_tokens_per_session: Number(row.avg_tokens_per_session),
		largest_session_tokens: Number(row.largest_session_tokens),
		sessions_over_100k_tokens: Number(row.sessions_over_100k_tokens)
	};
}

export type SessionDetail = {
	session_id: string;
	project_name: string;
	project_path: string;
	first_message_at: string;
	last_message_at: string;
	message_count: number;
	user_message_count: number;
	assistant_message_count: number;
};

export async function getSessionById(sessionId: string): Promise<SessionDetail | null> {
	const connection = await getConnection();
	const result = await connection.query(`
		SELECT
			session_id,
			project_name,
			project_path,
			first_message_at,
			last_message_at,
			CAST(message_count AS INTEGER) as message_count,
			CAST(user_message_count AS INTEGER) as user_message_count,
			CAST(assistant_message_count AS INTEGER) as assistant_message_count
		FROM sessions
		WHERE session_id = '${sessionId}'
		LIMIT 1
	`);
	const rows = result.toArray();
	if (rows.length === 0) return null;
	const row = rows[0];
	return {
		session_id: String(row.session_id),
		project_name: String(row.project_name),
		project_path: String(row.project_path),
		first_message_at: String(row.first_message_at),
		last_message_at: String(row.last_message_at),
		message_count: Number(row.message_count),
		user_message_count: Number(row.user_message_count),
		assistant_message_count: Number(row.assistant_message_count)
	};
}

export type MessageDetail = {
	id: string;
	type: string;
	timestamp: string;
	content: string;
	content_type: string;
	tool_name: string | null;
};

export async function getSessionMessages(sessionId: string): Promise<MessageDetail[]> {
	const connection = await getConnection();
	const result = await connection.query(`
		SELECT
			id,
			type,
			timestamp,
			content,
			COALESCE(content_type, '') as content_type,
			tool_name
		FROM messages
		WHERE session_id = '${sessionId}'
		ORDER BY timestamp ASC
	`);
	return result.toArray().map((row) => ({
		id: String(row.id),
		type: String(row.type),
		timestamp: String(row.timestamp),
		content: String(row.content),
		content_type: String(row.content_type),
		tool_name: row.tool_name ? String(row.tool_name) : null
	}));
}

export async function getSessionToolCalls(sessionId: string): Promise<ToolCount[]> {
	const connection = await getConnection();
	const result = await connection.query(`
		SELECT
			tool_name,
			CAST(COUNT(*) AS INTEGER) as count
		FROM tool_calls
		WHERE session_id = '${sessionId}'
		GROUP BY tool_name
		ORDER BY count DESC
	`);
	return result.toArray().map((row) => ({
		tool_name: String(row.tool_name),
		count: Number(row.count)
	}));
}
