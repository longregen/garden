--
-- PostgreSQL database dump
--

-- Dumped from database version 17.7
-- Dumped by pg_dump version 17.7

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: postgres
--

-- *not* creating schema, since initdb creates it


ALTER SCHEMA public OWNER TO postgres;

--
-- Name: fuzzystrmatch; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS fuzzystrmatch WITH SCHEMA public;


--
-- Name: EXTENSION fuzzystrmatch; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION fuzzystrmatch IS 'determine similarities and distance between strings';


--
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


--
-- Name: unaccent; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS unaccent WITH SCHEMA public;


--
-- Name: EXTENSION unaccent; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION unaccent IS 'text search dictionary that removes accents';


--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- Name: vector; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS vector WITH SCHEMA public;


--
-- Name: EXTENSION vector; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION vector IS 'vector data type and ivfflat and hnsw access methods';


--
-- Name: add_contact_tag(uuid, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.add_contact_tag(_contact_id uuid, _tagname text) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    _tag_id UUID;
BEGIN
    -- Find or create the tag
    INSERT INTO contact_tagnames (name)
    VALUES (_tagname)
    ON CONFLICT (name) DO NOTHING
    RETURNING tag_id INTO _tag_id;

    -- If the tag is new, get its ID
    IF NOT FOUND THEN
        SELECT tag_id INTO _tag_id FROM contact_tagnames WHERE name = _tagname;
    END IF;

    -- Create a contact_tag relation if not exists
    INSERT INTO contact_tags (contact_id, tag_id)
    SELECT _contact_id, _tag_id
    WHERE NOT EXISTS (
        SELECT 1 FROM contact_tags WHERE contact_id = _contact_id AND tag_id = _tag_id
    );
END;
$$;


ALTER FUNCTION public.add_contact_tag(_contact_id uuid, _tagname text) OWNER TO postgres;

--
-- Name: add_message_to_conversation(uuid, interval); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.add_message_to_conversation(IN _message_id uuid, IN time_gap interval)
    LANGUAGE plpgsql
    AS $$
DECLARE
    current_message messages%ROWTYPE;
    first_conversation_date TIMESTAMP;
    last_conversation_id UUID;
BEGIN
    -- Fetch the message details
    SELECT * INTO current_message FROM messages WHERE messages.message_id = _message_id;

    -- Find the last conversation's ID and the timestamp of its last message for the same room as the current message
    SELECT id, first_date_time, last_date_time
    INTO last_conversation_id, first_conversation_date
    FROM Conversations
    WHERE room_id = current_message.room_id
    ORDER BY last_date_time DESC
    LIMIT 1;

    -- Check if the last conversation for that room is within the time_gap
    IF first_conversation_date IS NOT NULL AND current_message.created_at - first_conversation_date <= time_gap THEN
        -- If within time_gap, add the message to the last conversation
        INSERT INTO Conversation_Message (conversation_id, message_id) VALUES (last_conversation_id, current_message.message_id);
    ELSE
        -- If not within time_gap or no previous conversation, create a new conversation
        last_conversation_id := uuid_generate_v4();
        INSERT INTO Conversations (id, room_id, first_message_id, first_date_time) VALUES (last_conversation_id, current_message.room_id, current_message.message_id, current_message.created_at);
        INSERT INTO Conversation_Message (conversation_id, message_id) VALUES (last_conversation_id, current_message.message_id);
    END IF;

    UPDATE Conversations SET 
        last_message_id = _message_id,
        last_date_time = current_message.created_at
    WHERE Conversations.id = last_conversation_id;
END;
$$;


ALTER PROCEDURE public.add_message_to_conversation(IN _message_id uuid, IN time_gap interval) OWNER TO postgres;

--
-- Name: add_message_to_session(uuid, interval); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.add_message_to_session(IN _message_id uuid, IN time_gap interval)
    LANGUAGE plpgsql
    AS $$
DECLARE
    current_message messages%ROWTYPE;
    first_session_date TIMESTAMP;
    last_session_id UUID;
    existing_entry BOOLEAN;
BEGIN
    -- Fetch the message details
    SELECT * INTO current_message FROM messages WHERE messages.message_id = _message_id;

    -- Find the last session's ID and the timestamp of its last message for the same room as the current message
    SELECT session_id, first_date_time, last_date_time
    INTO last_session_id, first_session_date
    FROM sessions
    WHERE room_id = current_message.room_id
    ORDER BY last_date_time DESC
    LIMIT 1;

    -- Check if the last session for that room is within the time_gap
    IF first_session_date IS NOT NULL AND current_message.event_datetime - first_session_date <= time_gap THEN
        -- Check if message is already in this session
        SELECT EXISTS(
            SELECT 1 FROM session_message 
            WHERE session_id = last_session_id AND message_id = current_message.message_id
        ) INTO existing_entry;
        
        -- If within time_gap and not already in session, add the message to the last session
        IF NOT existing_entry THEN
            INSERT INTO session_message (session_id, message_id) 
            VALUES (last_session_id, current_message.message_id);
        END IF;
    ELSE
        -- If not within time_gap or no previous session, create a new session
        last_session_id := gen_random_uuid();
        INSERT INTO sessions (session_id, room_id, first_message_id, first_date_time) 
        VALUES (last_session_id, current_message.room_id, current_message.message_id, current_message.event_datetime);
        
        -- Check if message is already in this session (unlikely but for safety)
        SELECT EXISTS(
            SELECT 1 FROM session_message 
            WHERE session_id = last_session_id AND message_id = current_message.message_id
        ) INTO existing_entry;
        
        IF NOT existing_entry THEN
            INSERT INTO session_message (session_id, message_id) 
            VALUES (last_session_id, current_message.message_id);
        END IF;
    END IF;

    UPDATE sessions SET 
        last_message_id = _message_id,
        last_date_time = current_message.event_datetime
    WHERE sessions.session_id = last_session_id;
END;
$$;


ALTER PROCEDURE public.add_message_to_session(IN _message_id uuid, IN time_gap interval) OWNER TO postgres;

--
-- Name: add_raw_message(text, json, timestamp without time zone); Type: PROCEDURE; Schema: public; Owner: gardener
--

CREATE PROCEDURE public.add_raw_message(IN raw_id text, IN input_json json, IN input_date timestamp without time zone)
    LANGUAGE plpgsql
    AS $$
DECLARE
    existent_ids INT;
BEGIN
    SELECT COUNT(*) INTO existent_ids FROM raw_messages r WHERE r.external_id = raw_id;

    IF existent_ids != 0 THEN
        RAISE EXCEPTION 'Message with ID % already found', raw_id;
    END IF;
    INSERT INTO raw_messages (external_id, content, created_at) VALUES (raw_id, input_json, input_date);

END;
$$;


ALTER PROCEDURE public.add_raw_message(IN raw_id text, IN input_json json, IN input_date timestamp without time zone) OWNER TO gardener;

--
-- Name: add_tag(uuid, text); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.add_tag(_item_id uuid, _tagname text) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    _tag_id UUID;
BEGIN
    -- Find or create the tag
    INSERT INTO tags (name)
    VALUES (_tagname)
    ON CONFLICT (name) DO NOTHING
    RETURNING id INTO _tag_id;

    -- If the tag is new, get its ID
    IF NOT FOUND THEN
        SELECT id INTO _tag_id FROM tags WHERE name = _tagname;
    END IF;

    -- Create an item_tag relation if not exists
    INSERT INTO item_tags (item_id, tag_id)
    SELECT _item_id, _tag_id
    WHERE NOT EXISTS (
        SELECT 1 FROM item_tags WHERE item_id = _item_id AND tag_id = _tag_id
    );
END;
$$;


ALTER FUNCTION public.add_tag(_item_id uuid, _tagname text) OWNER TO gardener;

--
-- Name: create_contact_entity(); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.create_contact_entity() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    _entity_id UUID;
BEGIN
    -- Create person entity
    INSERT INTO entities (name, type, created_at)
    VALUES (
        NEW.name,
        'person',
        NEW.creation_date
    )
    RETURNING entity_id INTO _entity_id;
    
    -- Create relationship
    INSERT INTO entity_relationships (entity_id, related_type, related_id, relationship_type)
    VALUES (_entity_id, 'contact', NEW.contact_id, 'identity');
    
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.create_contact_entity() OWNER TO gardener;

--
-- Name: create_item_with_tags(text, text, text); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.create_item_with_tags(title text, content text, tags text) RETURNS uuid
    LANGUAGE plpgsql
    AS $$
DECLARE
    new_item_id UUID;
    tag_name TEXT;
    tag_id UUID;
BEGIN
    -- Insert the new item and get its ID
    INSERT INTO items (title, contents)
    VALUES (title, content)
    RETURNING id INTO new_item_id;

    -- Split the tags string and loop through each tag
    FOREACH tag_name IN ARRAY STRING_TO_ARRAY(tags, ' ')
    LOOP
        -- Find or create the tag
        INSERT INTO tags (name)
        VALUES (tag_name)
        ON CONFLICT (name) DO NOTHING
        RETURNING id INTO tag_id;

        -- If the tag is new, get its ID
        IF NOT FOUND THEN
            SELECT id INTO tag_id FROM tags WHERE name = tag_name;
        END IF;

        -- Create an item_tag relation
        INSERT INTO item_tags (item_id, tag_id)
        VALUES (new_item_id, tag_id);
    END LOOP;

    -- Return the new item's ID
    RETURN new_item_id;
END;
$$;


ALTER FUNCTION public.create_item_with_tags(title text, content text, tags text) OWNER TO gardener;

--
-- Name: create_room_entity(); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.create_room_entity() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    _participant_count INTEGER;
    _entity_id UUID;
BEGIN
    -- Count participants
    SELECT COUNT(*) INTO _participant_count 
    FROM room_participants 
    WHERE room_id = NEW.room_id;
    
    -- Only create entities for rooms with more than 3 participants
    IF _participant_count > 3 THEN
        -- Create group_chat entity
        INSERT INTO entities (name, type, description, created_at)
        VALUES (
            COALESCE(NEW.display_name, NEW.user_defined_name, 'Unnamed Group'),
            'group_chat',
            '',
            COALESCE(NEW.last_activity, CURRENT_TIMESTAMP)
        )
        RETURNING entity_id INTO _entity_id;
        
        -- Create relationship
        INSERT INTO entity_relationships (entity_id, related_type, related_id, relationship_type)
        VALUES (_entity_id, 'room', NEW.room_id, 'identity');
    END IF;
    
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.create_room_entity() OWNER TO gardener;

--
-- Name: delete_category_source(uuid); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.delete_category_source(id_param uuid) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    DELETE FROM category_sources
    WHERE categories.category_id = id_param;
END;
$$;


ALTER FUNCTION public.delete_category_source(id_param uuid) OWNER TO gardener;

--
-- Name: delete_contact_entity(); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.delete_contact_entity() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    _entity_id UUID;
BEGIN
    -- Find the entity associated with this contact
    SELECT entity_id INTO _entity_id
    FROM entity_relationships
    WHERE related_type = 'contact' AND related_id = OLD.contact_id AND relationship_type = 'identity';
    
    IF _entity_id IS NOT NULL THEN
        -- Soft delete the entity
        UPDATE entities
        SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
        WHERE entity_id = _entity_id;
        
        -- Delete relationships
        DELETE FROM entity_relationships
        WHERE entity_id = _entity_id;
    END IF;
    
    RETURN OLD;
END;
$$;


ALTER FUNCTION public.delete_contact_entity() OWNER TO gardener;

--
-- Name: encode_zbase32(bigint); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.encode_zbase32(num bigint) RETURNS text
    LANGUAGE plpgsql
    AS $$
DECLARE
    -- Define the zbase32 character set
    chars CHAR[32] := ARRAY[
        'y', 'b', 'n', 'd', 'r', 'f', 'g', '8',
        'e', 'j', 'k', 'm', 'c', 'p', 'q', 'x',
        'o', 't', '1', 'u', 'w', 'i', 's', 'z',
        'a', '3', '4', '5', 'h', '7', '6', '9'];
    result TEXT := '';
	val BIGINT;
    bit_group INTEGER;
BEGIN
	RAISE NOTICE 'value: %', num;
 	-- Encodes up to 60 bits
    FOR i IN 0..11 LOOP
		RAISE NOTICE 'counter: %, %, %', num, (1::bigint << ((i)*5)), (num >>i*5) & 31;
		val := (num + 1) >> i*5;
        bit_group := val & 31; -- 5 bits per group
        IF val > 0 THEN
        	result := chars[bit_group + 1] || result; -- +1 because arrays are 1-indexed in PostgreSQL
		END IF;
    END LOOP;
    RETURN result;
END;
$$;


ALTER FUNCTION public.encode_zbase32(num bigint) OWNER TO gardener;

--
-- Name: ensure_contact_exists(text, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ensure_contact_exists(p_matrix_user_id text, p_display_name text) RETURNS uuid
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_contact_id UUID;
    v_contact_name TEXT;
    v_source_id TEXT;
BEGIN
    -- Ensure we have a non-null source_id
    IF p_matrix_user_id IS NULL THEN
        RAISE EXCEPTION 'Matrix user ID cannot be NULL';
    END IF;
    
    v_source_id := p_matrix_user_id;
    
    -- Check if contact already exists via contact_sources
    SELECT contact_id INTO v_contact_id
    FROM contact_sources
    WHERE source_id = v_source_id;
    
    -- If not found, create new contact
    IF v_contact_id IS NULL THEN
        -- Ensure we have a non-null name
        v_contact_name := COALESCE(p_display_name, v_source_id, '(Unknown User)');
        
        INSERT INTO contacts (name, creation_date, last_update)
        VALUES (v_contact_name, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        RETURNING contact_id INTO v_contact_id;
        
        -- Add to contact_sources
        INSERT INTO contact_sources (contact_id, source_id, source_name)
        VALUES (v_contact_id, v_source_id, 'migration');
        
        -- Add to contact_known_names if display name exists
        IF p_display_name IS NOT NULL THEN
            INSERT INTO contact_known_names (contact_id, name)
            VALUES (v_contact_id, p_display_name);
        END IF;
    ELSE
        -- Update known names if new display name
        IF p_display_name IS NOT NULL THEN
            -- Check if this name already exists for the contact
            IF NOT EXISTS (
                SELECT 1 FROM contact_known_names 
                WHERE contact_id = v_contact_id AND name = p_display_name
            ) THEN
                INSERT INTO contact_known_names (contact_id, name)
                VALUES (v_contact_id, p_display_name);
            END IF;
        END IF;
    END IF;
    
    RETURN v_contact_id;
END;
$$;


ALTER FUNCTION public.ensure_contact_exists(p_matrix_user_id text, p_display_name text) OWNER TO postgres;

--
-- Name: ensure_room_exists(text, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ensure_room_exists(p_matrix_room_id text, p_room_display_name text) RETURNS uuid
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_room_id UUID;
    v_current_time TIMESTAMP := CURRENT_TIMESTAMP;
BEGIN
    -- Check if room already exists
    SELECT room_id INTO v_room_id
    FROM rooms
    WHERE source_id = p_matrix_room_id;
    
    -- If not found, create new room
    IF v_room_id IS NULL THEN
        INSERT INTO rooms (source_id, display_name, last_activity)
        VALUES (p_matrix_room_id, p_room_display_name, v_current_time)
        RETURNING room_id INTO v_room_id;
    ELSE
        -- Update room display name if provided
        IF p_room_display_name IS NOT NULL THEN
            UPDATE rooms
            SET display_name = p_room_display_name,
                last_activity = v_current_time
            WHERE room_id = v_room_id;
        END IF;
    END IF;
    
    -- Add to room_known_names if display name exists
    IF p_room_display_name IS NOT NULL THEN
        -- Check if this name already exists for the room
        IF NOT EXISTS (
            SELECT 1 FROM room_known_names 
            WHERE room_id = v_room_id AND name = p_room_display_name
        ) THEN
            INSERT INTO room_known_names (room_id, name, last_time)
            VALUES (v_room_id, p_room_display_name, v_current_time);
        ELSE
            -- Update last_time for existing name
            UPDATE room_known_names
            SET last_time = v_current_time
            WHERE room_id = v_room_id AND name = p_room_display_name;
        END IF;
    END IF;
    
    RETURN v_room_id;
END;
$$;


ALTER FUNCTION public.ensure_room_exists(p_matrix_room_id text, p_room_display_name text) OWNER TO postgres;

--
-- Name: ensure_room_participant(uuid, uuid, timestamp without time zone); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.ensure_room_participant(p_room_id uuid, p_contact_id uuid, p_event_time timestamp without time zone) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Check if participant already exists
    IF EXISTS (
        SELECT 1 FROM room_participants
        WHERE room_id = p_room_id AND contact_id = p_contact_id
    ) THEN
        -- Update last presence time
        UPDATE room_participants
        SET known_last_presence = p_event_time
        WHERE room_id = p_room_id AND contact_id = p_contact_id
        AND (known_last_presence IS NULL OR known_last_presence < p_event_time);
    ELSE
        -- Add new participant
        INSERT INTO room_participants (room_id, contact_id, known_last_presence)
        VALUES (p_room_id, p_contact_id, p_event_time);
    END IF;
END;
$$;


ALTER FUNCTION public.ensure_room_participant(p_room_id uuid, p_contact_id uuid, p_event_time timestamp without time zone) OWNER TO postgres;

--
-- Name: filter_bookmarks(uuid, text, timestamp without time zone, timestamp without time zone); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.filter_bookmarks(category_id_input uuid, search_query text, start_creation_date timestamp without time zone, end_creation_date timestamp without time zone) RETURNS TABLE(bookmark_id uuid, url text, title text, creation_date timestamp without time zone, category_id uuid, category_name text)
    LANGUAGE plpgsql
    AS $_$
BEGIN
    RETURN QUERY
    SELECT b.bookmark_id, b.url, t.title, b.creation_date, c.category_id, c.name
    FROM bookmarks b
    JOIN bookmark_category bc ON b.bookmark_id = bc.bookmark_id
    JOIN bookmark_titles t ON t.bookmark_id = b.bookmark_id
    JOIN categories c ON bc.category_id = c.category_id
    WHERE ($1 IS NULL OR c.category_id = $1)
    AND ($2 IS NULL OR b.url LIKE '%' || $2 || '%')
    AND (b.creation_date BETWEEN COALESCE($3, '-infinity') AND COALESCE($4, 'infinity'))
    ORDER BY b.creation_date DESC;
END;
$_$;


ALTER FUNCTION public.filter_bookmarks(category_id_input uuid, search_query text, start_creation_date timestamp without time zone, end_creation_date timestamp without time zone) OWNER TO gardener;

--
-- Name: generate_conversations(interval); Type: PROCEDURE; Schema: public; Owner: gardener
--

CREATE PROCEDURE public.generate_conversations(IN time_gap interval)
    LANGUAGE plpgsql
    AS $$
DECLARE
    last_message messages%ROWTYPE;
    current_message messages%ROWTYPE;
    current_conversation_id UUID;
    is_last_message_assigned BOOLEAN := FALSE;
BEGIN
    -- Iterate through messages in chronological order
    FOR current_message IN SELECT * FROM messages ORDER BY room_id, created_at LOOP
        -- Check if this is the first message in the loop or if it belongs to a new room
        IF NOT is_last_message_assigned OR last_message.room_id <> current_message.room_id THEN
            -- Start a new conversation
            current_conversation_id := uuid_generate_v4();
            INSERT INTO Conversations (id, room_id, first_date_time) VALUES (current_conversation_id, current_message.room_id, current_message.created_at);
            is_last_message_assigned := TRUE;
        ELSIF is_last_message_assigned AND current_message.created_at - last_message.created_at > time_gap THEN
            -- If the current message is beyond the time_gap, start a new conversation
            current_conversation_id := uuid_generate_v4();
            INSERT INTO Conversations (id, room_id, first_date_time) VALUES (current_conversation_id, current_message.room_id, current_message.created_at);
        END IF;
        
        -- Add current message to the current conversation
        INSERT INTO Conversation_Message (conversation_id, message_id) VALUES (current_conversation_id, current_message.message_id);
        
        -- Update last_message
        last_message := current_message;
    END LOOP;
END;
$$;


ALTER PROCEDURE public.generate_conversations(IN time_gap interval) OWNER TO gardener;

--
-- Name: generate_random_id(text); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.generate_random_id(prefix text) RETURNS text
    LANGUAGE plpgsql
    AS $$
DECLARE
    random_64bit bigint;
    encoded_id text;
BEGIN
    -- Generate a random 64-bit integer
    -- random() returns a double precision value in [0, 1)
    -- We scale it to the range of bigint by multiplying by 2^63 and casting
    random_64bit := (random() * 1152921504606846976)::bigint;

    -- Encode this random number using zbase32
    encoded_id := encode_zbase32(random_64bit);

    -- Return the concatenated result with the prefix
    RETURN prefix || ':' || encoded_id;
END;
$$;


ALTER FUNCTION public.generate_random_id(prefix text) OWNER TO gardener;

--
-- Name: get_all_contact_tagnames(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_all_contact_tagnames() RETURNS TABLE(tag_id uuid, name text)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT ctn.tag_id, ctn.name
    FROM contact_tagnames ctn
    ORDER BY ctn.name;
END;
$$;


ALTER FUNCTION public.get_all_contact_tagnames() OWNER TO postgres;

--
-- Name: get_category_by_id(uuid); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.get_category_by_id(category_id_param uuid) RETURNS TABLE(category_id uuid, name text)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT categories.category_id, categories.name
    FROM categories
    WHERE categories.category_id = category_id_param;
END;
$$;


ALTER FUNCTION public.get_category_by_id(category_id_param uuid) OWNER TO gardener;

--
-- Name: get_contact_tags(uuid); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_contact_tags(contact_id uuid) RETURNS TABLE(tag_id uuid, name text)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT ct.tag_id, ctn.name
    FROM contact_tags ct
    JOIN contact_tagnames ctn ON ct.tag_id = ctn.tag_id
    WHERE ct.contact_id = get_contact_tags.contact_id
    ORDER BY ctn.name;
END;
$$;


ALTER FUNCTION public.get_contact_tags(contact_id uuid) OWNER TO postgres;

--
-- Name: get_conversations(uuid); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.get_conversations(_conversation_id uuid) RETURNS jsonb
    LANGUAGE plpgsql
    AS $$
DECLARE
    message_ids UUID[];
    v_summary TEXT := '';
    v_messages JSONB := '[]';
    v_transcriptions JSONB := '{}';
    v_bookmarks JSONB := '{}';
    v_conversations JSONB;
BEGIN
    -- Retrieve recent conversations
    SELECT JSONB_AGG(t) INTO v_conversations FROM (
        SELECT * FROM conversations
        LEFT JOIN rooms ON rooms.room_id = conversations.room_id
        ORDER BY last_date_time DESC LIMIT 10
    ) t;

    -- Retrieve conversation summary if a specific conversation_id is provided
    IF _conversation_id IS NOT NULL THEN
        SELECT summary INTO v_summary FROM conversation_summaries
        WHERE _conversation_id = conversation_id;

        SELECT array_agg(cm.message_id) INTO message_ids
        FROM conversation_message cm
        WHERE cm.conversation_id = _conversation_id;

        -- Handle bookmarks
        WITH bookmarked_ids AS (
          SELECT bookmark_id FROM bookmark_sources
            WHERE source_uri IN (
              SELECT external_id FROM messages m where m.message_id = ANY(message_ids)
            )
        )
        SELECT JSONB_OBJECT_AGG(cs.source_uri, jsonb_build_object(
            'bookmark_id', b.bookmark_id,
            'title', ts.title,
            'lynx', lynx.processed_content,
            'summary', pcs.content,
            'reader', reader.processed_content
        )) INTO v_bookmarks
        FROM bookmarks b
        LEFT JOIN bookmark_titles ts ON b.bookmark_id = ts.bookmark_id
        
        LEFT JOIN bookmark_sources cs ON b.bookmark_id = cs.bookmark_id
        LEFT JOIN processed_contents lynx ON b.bookmark_id = lynx.bookmark_id AND lynx.strategy_used = 'lynx'
        LEFT JOIN processed_contents reader ON b.bookmark_id = reader.bookmark_id AND reader.strategy_used = 'reader'
        LEFT JOIN bookmark_content_references pcs ON b.bookmark_id = pcs.bookmark_id AND pcs.strategy = 'summary-reader'
        WHERE b.bookmark_id IN (SELECT * FROM bookmarked_ids);

        -- Handle transcriptions
        WITH transcribed_messages AS (
            SELECT message_id
            FROM messages m 
            WHERE m.content->>'msgtype' = 'm.audio'
              AND m.message_id = ANY(message_ids)
        )
        SELECT JSONB_OBJECT_AGG(source, data) INTO v_transcriptions FROM observations
        WHERE type = 'transcription' AND ref = ANY(ARRAY(SELECT message_id FROM transcribed_messages));

        -- Retrieve messages for the specific conversation
        SELECT JSONB_AGG(x) FROM (
           SELECT jsonb_build_object(
                'id', m.message_id, 
                'from', m.sender_id, 
                'content', m.content, 
                'external_id', m.external_id, 
                'created_at', m.created_at,
                'text', CASE WHEN m.content->>'msgtype' = 'm.text' THEN jsonb_build_object(
                        'body', m.content->>'body',
                        'bookmark', jsonb_extract_path(v_bookmarks, m.external_id)
                    ) ELSE NULL END,
                'audio', CASE WHEN m.content->>'msgtype' = 'm.audio' THEN jsonb_build_object(
                        'id', m.message_id,
                        'body', m.content->>'body',
                        'transcription', jsonb_extract_path(v_transcriptions, 'raw_messages:' || m.external_id, 'text')
                    ) ELSE NULL END,
                'notice', CASE WHEN m.content->>'msgtype' = 'm.notice' THEN jsonb_build_object(
                        'body', m.content->>'body'
                    ) ELSE NULL END,
                'image', CASE WHEN m.content->>'msgtype' = 'm.image' THEN jsonb_build_object(
                        'id', m.external_id
                    ) ELSE NULL END
            ) as x
            FROM messages m
            WHERE m.message_id = ANY(message_ids)
            ORDER BY m.created_at DESC
        ) as y INTO v_messages;

    END IF;
    
    -- Compile the final result JSON
    RETURN jsonb_build_object(
        'conversations', v_conversations,
        'summary', v_summary,
        'messages', v_messages
    );
END;
$$;


ALTER FUNCTION public.get_conversations(_conversation_id uuid) OWNER TO gardener;

--
-- Name: get_sessions(uuid); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_sessions(_session_id uuid) RETURNS jsonb
    LANGUAGE plpgsql
    AS $$
DECLARE
    message_ids UUID[];
    v_summary TEXT := '';
    v_messages JSONB := '[]';
    v_transcriptions JSONB := '{}';
    v_bookmarks JSONB := '{}';
    v_sessions JSONB;
BEGIN
    -- Retrieve recent sessions
    SELECT JSONB_AGG(t) INTO v_sessions FROM (
        SELECT * FROM sessions
        LEFT JOIN rooms ON rooms.room_id = sessions.room_id
        ORDER BY last_date_time DESC LIMIT 10
    ) t;

    -- Retrieve session summary if a specific session_id is provided
    IF _session_id IS NOT NULL THEN
        SELECT summary INTO v_summary FROM session_summaries
        WHERE _session_id = session_id;

        SELECT array_agg(sm.message_id) INTO message_ids
        FROM session_message sm
        WHERE sm.session_id = _session_id;

        -- Handle bookmarks
        WITH bookmarked_ids AS (
          SELECT bookmark_id FROM bookmark_sources
            WHERE source_uri IN (
              SELECT event_id FROM messages m where m.message_id = ANY(message_ids)
            )
        )
        SELECT JSONB_OBJECT_AGG(cs.source_uri, jsonb_build_object(
            'bookmark_id', b.bookmark_id,
            'title', ts.title,
            'lynx', lynx.processed_content,
            'summary', pcs.content,
            'reader', reader.processed_content
        )) INTO v_bookmarks
        FROM bookmarks b
        LEFT JOIN bookmark_titles ts ON b.bookmark_id = ts.bookmark_id
        
        LEFT JOIN bookmark_sources cs ON b.bookmark_id = cs.bookmark_id
        LEFT JOIN processed_contents lynx ON b.bookmark_id = lynx.bookmark_id AND lynx.strategy_used = 'lynx'
        LEFT JOIN processed_contents reader ON b.bookmark_id = reader.bookmark_id AND reader.strategy_used = 'reader'
        LEFT JOIN bookmark_content_references pcs ON b.bookmark_id = pcs.bookmark_id AND pcs.strategy = 'summary-reader'
        WHERE b.bookmark_id IN (SELECT * FROM bookmarked_ids);

        -- Handle transcriptions
        WITH transcribed_messages AS (
            SELECT message_id
            FROM messages m 
            WHERE m.msgtype = 'm.audio'
              AND m.message_id = ANY(message_ids)
        )
        SELECT JSONB_OBJECT_AGG(source, data) INTO v_transcriptions FROM observations
        WHERE type = 'transcription' AND ref = ANY(ARRAY(SELECT message_id FROM transcribed_messages));

        -- Retrieve messages for the specific session
        SELECT JSONB_AGG(x) FROM (
           SELECT jsonb_build_object(
                'id', m.message_id, 
                'from', m.sender_contact_id, 
                'event_id', m.event_id,
                'body', m.body,
                'formatted_body', m.formatted_body,
                'format', m.format,
                'msgtype', m.msgtype,
                'message_type', m.message_type,
                'message_classification', m.message_classification,
                'is_edited', m.is_edited,
                'is_reply', m.is_reply,
                'reply_to_event_id', m.reply_to_event_id,
                'created_at', m.event_datetime,
                'text', CASE WHEN m.msgtype = 'm.text' THEN jsonb_build_object(
                        'body', m.body,
                        'bookmark', jsonb_extract_path(v_bookmarks, m.event_id)
                    ) ELSE NULL END,
                'audio', CASE WHEN m.msgtype = 'm.audio' THEN jsonb_build_object(
                        'id', m.message_id,
                        'body', m.body,
                        'transcription', jsonb_extract_path(v_transcriptions, 'raw_messages:' || m.event_id, 'text')
                    ) ELSE NULL END,
                'notice', CASE WHEN m.msgtype = 'm.notice' THEN jsonb_build_object(
                        'body', m.body
                    ) ELSE NULL END,
                'image', CASE WHEN m.msgtype = 'm.image' THEN jsonb_build_object(
                        'id', m.event_id
                    ) ELSE NULL END
            ) as x
            FROM messages m
            WHERE m.message_id = ANY(message_ids)
            ORDER BY m.event_datetime DESC
        ) as y INTO v_messages;

    END IF;
    
    -- Compile the final result JSON
    RETURN jsonb_build_object(
        'sessions', v_sessions,
        'summary', v_summary,
        'messages', v_messages
    );
END;
$$;


ALTER FUNCTION public.get_sessions(_session_id uuid) OWNER TO postgres;

--
-- Name: ident(text, bigint); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.ident(prefix text, num bigint) RETURNS text
    LANGUAGE plpgsql
    AS $$
DECLARE
    encoded_value text;
BEGIN
    -- Encode the number using the zbase32 function
    encoded_value := encode_zbase32(num);

    -- Return the concatenated result
    RETURN prefix || ':' || encoded_value;
END;
$$;


ALTER FUNCTION public.ident(prefix text, num bigint) OWNER TO gardener;

--
-- Name: merge_contacts(uuid, uuid); Type: PROCEDURE; Schema: public; Owner: gardener
--

CREATE PROCEDURE public.merge_contacts(IN source_contact_id uuid, IN target_contact_id uuid)
    LANGUAGE plpgsql
    AS $$
DECLARE
    source_entity_id UUID;
    target_entity_id UUID;
    ref_record RECORD;
    source_importance INTEGER;
    source_closeness INTEGER;
    source_fondness INTEGER;
    target_has_evals BOOLEAN;
BEGIN
    -- Transfer contact sources from source to target
    UPDATE contact_sources
    SET contact_id = target_contact_id
    WHERE contact_id = source_contact_id;

    UPDATE messages
    SET sender_contact_id = target_contact_id
    WHERE sender_contact_id = source_contact_id;

    -- Transfer known names from source to target
    UPDATE contact_known_names r
    SET contact_id = target_contact_id
    WHERE contact_id = source_contact_id
    AND NOT EXISTS (
        select 1 from contact_known_names WHERE contact_id = r.contact_id AND name = r.name
    );
    DELETE from contact_known_names where contact_id = source_contact_id;

    -- Transfer known avatars from source to target
    UPDATE contact_known_avatars r
    SET contact_id = target_contact_id
    WHERE contact_id = source_contact_id
    AND NOT EXISTS (
        select 1 from contact_known_avatars WHERE contact_id = r.contact_id AND avatar = r.avatar
    );
    DELETE from contact_known_avatars where contact_id = source_contact_id;

    -- Transfer known room participants
    UPDATE room_participants
    SET contact_id = target_contact_id
    WHERE contact_id = source_contact_id;
    
    -- Transfer contact tags (if any)
    -- First, get all tag_ids from source contact
    WITH source_tags AS (
        SELECT tag_id FROM contact_tags WHERE contact_id = source_contact_id
    )
    -- Then insert them for target contact if they don't already exist
    INSERT INTO contact_tags (contact_id, tag_id)
    SELECT target_contact_id, st.tag_id
    FROM source_tags st
    WHERE NOT EXISTS (
        SELECT 1 FROM contact_tags 
        WHERE contact_id = target_contact_id AND tag_id = st.tag_id
    );
    -- Delete source contact tags
    DELETE FROM contact_tags WHERE contact_id = source_contact_id;
    
    -- Transfer contact evaluations (if any)
    -- Get source evaluations
    SELECT importance, closeness, fondness INTO source_importance, source_closeness, source_fondness
    FROM contact_evals WHERE contact_id = source_contact_id;
    
    -- Check if target has evaluations
    SELECT EXISTS(SELECT 1 FROM contact_evals WHERE contact_id = target_contact_id) INTO target_has_evals;
    
    IF FOUND AND source_importance IS NOT NULL THEN
        IF target_has_evals THEN
            -- Update target evaluations with non-null values from source
            UPDATE contact_evals
            SET 
                importance = COALESCE(contact_evals.importance, source_importance),
                closeness = COALESCE(contact_evals.closeness, source_closeness),
                fondness = COALESCE(contact_evals.fondness, source_fondness),
                updated_at = NOW()
            WHERE contact_id = target_contact_id;
        ELSE
            -- Insert new evaluation record for target
            INSERT INTO contact_evals (contact_id, importance, closeness, fondness)
            VALUES (target_contact_id, source_importance, source_closeness, source_fondness);
        END IF;
        
        -- Delete source evaluations
        DELETE FROM contact_evals WHERE contact_id = source_contact_id;
    END IF;
    
    -- Find entities connected to the source and target contacts
    SELECT entity_id INTO source_entity_id
    FROM entity_relationships
    WHERE related_type = 'contact' 
    AND related_id = source_contact_id 
    AND relationship_type = 'identity';
    
    SELECT entity_id INTO target_entity_id
    FROM entity_relationships
    WHERE related_type = 'contact' 
    AND related_id = target_contact_id 
    AND relationship_type = 'identity';
    
    -- If both contacts have associated entities, merge them
    IF source_entity_id IS NOT NULL AND target_entity_id IS NOT NULL THEN
        -- Update entity references to point to the target entity
        UPDATE entity_references
        SET entity_id = target_entity_id
        WHERE entity_id = source_entity_id;
        
        -- Update entity relationships to point to the target entity
        UPDATE entity_relationships
        SET entity_id = target_entity_id
        WHERE entity_id = source_entity_id
        AND NOT (related_type = 'contact' AND related_id = source_contact_id AND relationship_type = 'identity');
        
        -- For each entity reference, update the text content in the source
        FOR ref_record IN 
            SELECT id, source_type, source_id, reference_text 
            FROM entity_references 
            WHERE entity_id = target_entity_id
        LOOP
            -- Update the content based on the source type
            CASE ref_record.source_type
                WHEN 'item' THEN
                    -- Update references by UUID
                    -- Format: [[entity-uuid]] or [[alias][entity-uuid]]
                    UPDATE items
                    SET contents = REGEXP_REPLACE(
                        contents,
                        '\[\[([^\]]*)\]\[' || source_entity_id || '\]\]',
                        '[[\1][' || target_entity_id || ']]',
                        'g'
                    )
                    WHERE id = ref_record.source_id::uuid;
                    
                    -- Also handle the simple case: [[entity-uuid]]
                    UPDATE items
                    SET contents = REGEXP_REPLACE(
                        contents,
                        '\[\[' || source_entity_id || '\]\]',
                        '[[' || target_entity_id || ']]',
                        'g'
                    )
                    WHERE id = ref_record.source_id::uuid;
                
                ELSE
                    -- Do nothing for other source types
                    NULL;
            END CASE;
        END LOOP;
        
        -- Soft delete the source entity
        UPDATE entities
        SET deleted_at = NOW(), updated_at = NOW()
        WHERE entity_id = source_entity_id;
    ELSIF source_entity_id IS NOT NULL AND target_entity_id IS NULL THEN
        -- If only source has an entity, update its relationship to point to target contact
        UPDATE entity_relationships
        SET related_id = target_contact_id
        WHERE entity_id = source_entity_id
        AND related_type = 'contact'
        AND related_id = source_contact_id
        AND relationship_type = 'identity';
    END IF;

    -- Update references in messages_old table if it exists
    BEGIN
        -- Check if messages_old table exists and update references
        UPDATE messages_old
        SET sender_id = target_contact_id
        WHERE sender_id = source_contact_id;
    EXCEPTION WHEN undefined_table THEN
        -- Table doesn't exist, do nothing
        NULL;
    END;
    
    -- Now try to delete the source contact
    -- If it still fails, we'll keep the contact but mark it as merged
    BEGIN
        DELETE FROM contacts
        WHERE contact_id = source_contact_id;
    EXCEPTION WHEN foreign_key_violation THEN
        -- If we can't delete due to foreign key constraints, 
        -- update the contact to indicate it's been merged
        UPDATE contacts
        SET 
            name = name || ' (merged to ' || target_contact_id || ')',
            notes = COALESCE(notes, '') || E'\nThis contact was merged into contact ' || target_contact_id || ' on ' || NOW()
        WHERE contact_id = source_contact_id;
    END;
END;
$$;


ALTER PROCEDURE public.merge_contacts(IN source_contact_id uuid, IN target_contact_id uuid) OWNER TO gardener;

--
-- Name: message_text_search_update(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.message_text_search_update() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.search_vector := to_tsvector('english', NEW.text_content);
    RETURN NEW;
END
$$;


ALTER FUNCTION public.message_text_search_update() OWNER TO postgres;

--
-- Name: notify_new_bookmark(); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.notify_new_bookmark() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    PERFORM pg_notify('new_bookmark', row_to_json(NEW)::text);
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.notify_new_bookmark() OWNER TO gardener;

--
-- Name: notify_new_item(); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.notify_new_item() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
	json TEXT;
BEGIN
    -- Define the JSON payload with only the fields "id", "title", "slug", and "created"
    json := json_build_object(
      'id', NEW.id,
      'title', NEW.title,
      'slug', NEW.slug,
	  'created', NEW.created
    );  
    -- Trigger the NOTIFY event with the JSON payload
    PERFORM pg_notify('new_item', json::text);
    
    -- Return the new row (required for INSERT triggers)
    RETURN NEW;
END; $$;


ALTER FUNCTION public.notify_new_item() OWNER TO gardener;

--
-- Name: parse_entity_references(text); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.parse_entity_references(content text) RETURNS TABLE(original_text text, entity_name text, display_text text)
    LANGUAGE plpgsql
    AS $$
DECLARE
    regex_pattern TEXT := '\[\[(.*?)(?:\]\[([^\]]+))?\]\]';
    matches TEXT[];
BEGIN
    FOR matches IN SELECT regexp_matches(content, regex_pattern, 'g') LOOP
        original_text := matches[0];
        entity_name := matches[1];
        display_text := COALESCE(matches[2], entity_name);
        RETURN NEXT;
    END LOOP;
END;
$$;


ALTER FUNCTION public.parse_entity_references(content text) OWNER TO gardener;

--
-- Name: populate_data_from_xml(xml); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.populate_data_from_xml(v_entry xml) RETURNS uuid
    LANGUAGE plpgsql
    AS $$
DECLARE
    v_paper_id UUID;
    v_author_id UUID;
    v_bibtex_id UUID;
    v_title TEXT;
    v_abstract TEXT;
    v_url TEXT;
    v_author_name TEXT;
    v_bibtex_data TEXT;
BEGIN
    -- Extract paper information
    v_title := (xpath('//feed/entry/title', v_entry))::text;
    v_abstract := (xpath('//feed/entry/summary', v_entry))::text;
    v_url := (xpath('//feed/entry/link/@href', v_entry))::text;

    -- Generate UUID for paper
    v_paper_id := uuid_generate_v4();

    -- Insert into papers table
    INSERT INTO papers (paper_id, title, abstract, url)
    VALUES (v_paper_id, v_title, v_abstract, v_url);

    -- Process each author
    FOR v_author_name IN SELECT (xpath('//feed/entry/author/name/text()', v_entry))[i]::text FROM generate_series(1, xpath_count(v_entry, '//feed/entry/author/name'))
    LOOP
        -- Generate UUID for author
        v_author_id := uuid_generate_v4();

        -- Insert into authors table
        INSERT INTO authors (author_id, name) VALUES (v_author_id, v_author_name);

        -- Link paper and author
        INSERT INTO paper_authors (paper_id, author_id) VALUES (v_paper_id, v_author_id);
    END LOOP;

    -- Insert bibtex data
    -- Assuming bibtex data is available and v_bibtex_data is set properly
    v_bibtex_id := uuid_generate_v4();
    INSERT INTO bibtex (bibtex_id, paper_id, bibtex_data) VALUES (v_bibtex_id, v_paper_id, v_bibtex_data);

    -- Return the paper ID
    RETURN v_paper_id;
END;
$$;


ALTER FUNCTION public.populate_data_from_xml(v_entry xml) OWNER TO gardener;

--
-- Name: process_chat_room_bookmark(json); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.process_chat_room_bookmark(IN input_json json)
    LANGUAGE plpgsql
    AS $$
DECLARE
    room_id TEXT;
    in_url TEXT;
    created TIMESTAMP;
    event_id TEXT;
    room_name TEXT;
    user_count INTEGER;
    cat_id UUID;
    book_id UUID;
BEGIN
    -- Extract fields from JSON
    room_id := input_json -> 'source' ->> 'room_id';
    in_url := input_json -> 'source' -> 'content' ->> 'body';
    created := TO_TIMESTAMP((input_json -> 'source' ->> 'origin_server_ts')::BIGINT / 1000);
    event_id := input_json -> 'source' ->> 'event_id';
    room_name := input_json -> 'room' ->> 'name';

    SELECT COUNT(*) INTO user_count FROM (SELECT json_object_keys(input_json -> 'room' -> 'users')) v;

    -- Check if URL starts with http
    IF NOT (in_url ~* '^\s*http') THEN
        RAISE EXCEPTION 'URL does not start with http';
    END IF;

    -- Check for existing category
    SELECT category_id INTO cat_id FROM category_sources WHERE source_uri = room_id;

    IF NOT FOUND THEN
        IF user_count = 1 THEN
            -- Insert new category and get cat_id
            INSERT INTO categories(name) VALUES (room_name) RETURNING category_id INTO cat_id;

            -- Create category source
            INSERT INTO category_sources(category_id, source_uri) VALUES (cat_id, room_id);
        ELSE
            SELECT category_id INTO cat_id FROM categories WHERE name = 'Shared With Me (Links)';
            IF NOT FOUND THEN
                INSERT INTO categories(name) VALUES ('Shared With Me (Links)') RETURNING category_id INTO cat_id;
            END IF;
            SELECT category_id INTO cat_id FROM category_sources WHERE source_uri = room_id;
            IF NOT FOUND THEN
                INSERT INTO category_sources(category_id, source_uri) VALUES (cat_id, room_id);
            END IF;
        END IF;
    END IF;

    -- Find if bookmark already exists in the category
    SELECT bookmarks.bookmark_id INTO book_id
      FROM bookmarks
      LEFT JOIN bookmark_category ON bookmarks.bookmark_id = bookmark_category.bookmark_id
      WHERE bookmark_category.category_id = cat_id
      AND bookmarks.url = TRIM(in_url);

    IF NOT FOUND THEN
      -- Find if bookmark url exists
      SELECT bookmarks.bookmark_id INTO book_id
        FROM bookmarks WHERE bookmarks.url = TRIM(in_url);

      IF NOT FOUND THEN
        -- Insert new bookmark and get bookmark_id
        INSERT INTO bookmarks(url, creation_date) VALUES (TRIM(in_url), created) RETURNING bookmark_id INTO book_id;
      END IF;

      -- Link bookmark with the category
      INSERT INTO bookmark_category(bookmark_id, category_id) VALUES (book_id, cat_id);

      -- Insert bookmark source
      INSERT INTO bookmark_sources (bookmark_id, source_uri, raw_source) VALUES (book_id, event_id, input_json);
    END IF;
END;
$$;


ALTER PROCEDURE public.process_chat_room_bookmark(IN input_json json) OWNER TO postgres;

--
-- Name: process_message(text); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.process_message(IN _id text)
    LANGUAGE plpgsql
    AS $$
DECLARE
  input_json JSONB;
BEGIN
  SELECT raw_messages.content INTO input_json FROM raw_messages WHERE external_id = _id;
  CALL process_message_body(input_json);
END;
$$;


ALTER PROCEDURE public.process_message(IN _id text) OWNER TO postgres;

--
-- Name: process_message_body(jsonb); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.process_message_body(IN input_json jsonb)
    LANGUAGE plpgsql
    AS $$
DECLARE
  _room_raw_id TEXT;
  _room_id UUID;
  _room_last_activity TIMESTAMP;
  _room_members JSONB;
  _room_current_name TEXT;
  _room_name TEXT;
  _room_avatar TEXT;
  _room_platform TEXT := 'matrix';

  _event_id TEXT;
  _sender_id UUID;
  _sender_external_id TEXT;
  _created TIMESTAMP;

  participant JSONB;
  _contact_source_external_id TEXT;
  _contact_source_id UUID;
  _contact_id UUID;
  _contact_name TEXT;
  _contact_avatar TEXT;
  _contact_source TEXT;
  _participant_count INTEGER := 0;
  _regular_participant_count INTEGER := 0;
  _message_body TEXT;
  _room_type TEXT;
  _msg_id UUID;

  -- Special contact IDs for bot detection
  _myself_contact_id CONSTANT UUID := '5a54e7ec-a4f4-42f8-881e-b0c41b910d97';
  _whatsapp_bot_id CONSTANT UUID := '67820976-64a2-48ff-bf9a-d5bb9f2d34ad';
  _signal_bot_id CONSTANT UUID := 'e2226b8a-327f-4692-80c3-d5048e34de08';
  _telegram_bot_id CONSTANT UUID := '5ed92a35-3b43-4c52-8011-907593659999';
BEGIN
  -- Extract fields from JSON
  _room_raw_id := input_json -> 'source' ->> 'room_id';
  _sender_external_id := input_json -> 'source' ->> 'sender';
  _room_name := input_json -> 'room' ->> 'name';
  _room_avatar := input_json -> 'room' ->> 'room_avatar_url';
  _created := TO_TIMESTAMP((input_json -> 'source' ->> 'origin_server_ts')::BIGINT / 1000);
  _room_members := input_json -> 'room' -> 'users';
  _event_id := input_json -> 'source' ->> 'event_id';

  IF _room_raw_id IS NULL THEN
      _room_raw_id := input_json -> 'room' ->> 'room_id';
  END IF;
  IF _room_raw_id IS NULL THEN
    RAISE EXCEPTION 'Problem! Message has null room. Details: %', input_json;
  END IF;

  IF _room_name IS NULL THEN
    _room_name := input_json ->> 'room_display_name';
  END IF;
  IF _room_name IS NULL THEN
    RAISE EXCEPTION 'Problem! Room has null name. Details: %', input_json;
  END IF;

  IF _sender_external_id IS NULL THEN
    _sender_external_id := input_json ->> 'few';
  END IF;
  IF _sender_external_id IS NULL THEN
    RAISE EXCEPTION 'Problem! Sender has null id. Details: %', input_json;
  END IF;

  -- Handle NULL room members
  IF _room_members IS NULL THEN
    _room_members := '{}'::jsonb;
  END IF;

  RAISE NOTICE 'Processing msg id: % room id: % sender id: %', _event_id, _room_raw_id, _sender_external_id;

  -- Use INSERT...ON CONFLICT for room upsert
  INSERT INTO rooms(room_id, display_name, source_id, last_activity)
  VALUES (uuid_generate_v4(), _room_name, _room_raw_id, _created)
  ON CONFLICT (source_id) DO UPDATE
    SET last_activity = GREATEST(rooms.last_activity, EXCLUDED.last_activity),
        display_name = CASE
          WHEN COALESCE(rooms.last_activity, 'epoch'::timestamp) < EXCLUDED.last_activity
          THEN EXCLUDED.display_name
          ELSE rooms.display_name
        END
  RETURNING room_id, last_activity, display_name INTO _room_id, _room_last_activity, _room_current_name;

  -- Initialize or update room_state
  INSERT INTO room_state(room_id, display_name, last_activity, room_type, room_platform)
  VALUES (_room_id, _room_name, _created, 'unknown', _room_platform)
  ON CONFLICT (room_id) DO UPDATE
    SET last_activity = GREATEST(room_state.last_activity, EXCLUDED.last_activity),
        display_name = CASE
          WHEN COALESCE(room_state.last_activity, 'epoch'::timestamp) < EXCLUDED.last_activity
          THEN EXCLUDED.display_name
          ELSE room_state.display_name
        END;

  -- Handle room name changes
  IF _room_current_name IS DISTINCT FROM _room_name AND COALESCE(_room_last_activity, 'epoch'::timestamp) < _created THEN
    INSERT INTO room_known_names(room_id, name, last_time)
    VALUES (_room_id, _room_name, _created) ON CONFLICT DO NOTHING;
  END IF;

  -- Handle room avatar
  IF _room_avatar IS NOT NULL THEN
    INSERT INTO room_known_avatars(room_id, avatar, earliest_date)
    VALUES(_room_id, _room_avatar, _created)
    ON CONFLICT (room_id, avatar)
    DO UPDATE SET
      earliest_date = LEAST(room_known_avatars.earliest_date, EXCLUDED.earliest_date);
  END IF;

  -- Process participants and detect platform
  FOR participant IN SELECT value FROM jsonb_each(_room_members) LOOP
    _contact_source_external_id := participant ->> 'user_id';
    _contact_name := participant ->> 'display_name';
    _participant_count := _participant_count + 1;

    IF _contact_name IS NULL THEN
      _contact_name := '(Unknown)';
    END IF;

    -- Check if contact source exists, create if not
    SELECT id, contact_id INTO _contact_source_id, _contact_id
    FROM contact_sources WHERE source_id = _contact_source_external_id;

    IF NOT FOUND THEN
      -- Create new contact first
      INSERT INTO contacts(contact_id, name)
      VALUES(uuid_generate_v4(), _contact_name)
      RETURNING contact_id INTO _contact_id;

      -- Then create contact source
      INSERT INTO contact_sources(id, contact_id, source_id, source_name)
      VALUES(uuid_generate_v4(), _contact_id, _contact_source_external_id, 'backfill')
      RETURNING id INTO _contact_source_id;
    END IF;

    -- Update contact known names and avatars
    IF _contact_name IS NOT NULL THEN
      INSERT INTO contact_known_names(contact_id, name)
      VALUES(_contact_id, _contact_name) ON CONFLICT DO NOTHING;
    END IF;

    _contact_avatar := participant ->> 'avatar_url';
    IF _contact_avatar IS NOT NULL THEN
      INSERT INTO contact_known_avatars(contact_id, avatar, earliest_date)
      VALUES(_contact_id, _contact_avatar, COALESCE(_created, CURRENT_TIMESTAMP))
      ON CONFLICT (contact_id, avatar)
      DO UPDATE SET
        earliest_date = LEAST(contact_known_avatars.earliest_date, EXCLUDED.earliest_date);
    END IF;

    -- Platform detection based on bot presence
    IF _contact_id = _whatsapp_bot_id THEN
      _room_platform := 'whatsapp';
    ELSIF _contact_id = _signal_bot_id THEN
      _room_platform := 'signal';
    ELSIF _contact_id = _telegram_bot_id THEN
      _room_platform := 'telegram';
    END IF;

    -- Count regular participants (non-bots and not myself)
    IF _contact_id NOT IN (_myself_contact_id, _whatsapp_bot_id, _signal_bot_id, _telegram_bot_id) THEN
      _regular_participant_count := _regular_participant_count + 1;
    END IF;

    -- Track sender
    IF _contact_source_external_id = _sender_external_id THEN
      _sender_id := _contact_id;
    END IF;
  END LOOP;

  -- Fallback sender lookup using correct sender external ID
  IF _sender_id IS NULL THEN
    SELECT contacts.contact_id INTO _sender_id
    FROM contacts NATURAL JOIN contact_sources
    WHERE contact_sources.source_id = _sender_external_id;
  END IF;

  -- Extract message text with proper handling for different types
  _message_body := CASE
    WHEN input_json->'source'->'content'->>'msgtype' = 'm.text' THEN
      input_json->'source'->'content'->>'body'
    WHEN input_json->'source'->'content'->>'msgtype' = 'm.image' THEN
      '📷 ' || COALESCE(input_json->'source'->'content'->>'body', 'Image')
    WHEN input_json->'source'->'content'->>'msgtype' = 'm.video' THEN
      '🎥 ' || COALESCE(input_json->'source'->'content'->>'body', 'Video')
    WHEN input_json->'source'->'content'->>'msgtype' = 'm.audio' THEN
      '🎵 ' || COALESCE(input_json->'source'->'content'->>'body', 'Audio')
    WHEN input_json->'source'->'content'->>'msgtype' = 'm.file' THEN
      '📎 ' || COALESCE(input_json->'source'->'content'->>'body', 'File')
    ELSE
      COALESCE(
        input_json->'source'->'content'->>'body',
        input_json->'source'->'content'->>'msgtype',
        '(No content)'
      )
  END;

  IF _sender_id IS NOT NULL THEN
    -- Insert message with duplicate handling
    INSERT INTO messages(
      event_id, 
      event_datetime,
      origin_server_ts,
      sender_contact_id, 
      room_id, 
      body,
      formatted_body,
      format,
      msgtype,
      created_at
    )
    VALUES (
      _event_id,
      _created,
      (input_json -> 'source' ->> 'origin_server_ts')::BIGINT,
      _sender_id, 
      _room_id, 
      input_json->'source'->'content'->>'body',
      input_json->'source'->'content'->>'formatted_body',
      input_json->'source'->'content'->>'format',
      input_json->'source'->'content'->>'msgtype',
      _created
    )
    ON CONFLICT (event_id) DO NOTHING
    RETURNING message_id INTO _msg_id;

    IF _msg_id IS NOT NULL THEN
      RAISE WARNING 'INSERTED MESSAGE';
    END IF;
  ELSE
    RAISE NOTICE 'Problem! Sender was not added as contact. Details: %', input_json;
  END IF;

  -- Update room participants for all contacts found
  FOR participant IN SELECT value FROM jsonb_each(_room_members) LOOP
    _contact_source_external_id := participant ->> 'user_id';

    -- Get contact_id
    SELECT contact_id INTO _contact_id
    FROM contact_sources
    WHERE source_id = _contact_source_external_id;

    IF _contact_id IS NOT NULL THEN
      -- Add room participant if not exists
      INSERT INTO room_participants (room_id, contact_id, known_last_presence)
      VALUES (_room_id, _contact_id, _created)
      ON CONFLICT (room_id, contact_id)
      DO UPDATE SET known_last_presence =
          CASE
              WHEN EXCLUDED.known_last_presence > room_participants.known_last_presence
              THEN EXCLUDED.known_last_presence
              ELSE room_participants.known_last_presence
          END;
    END IF;
  END LOOP;

  -- Determine room type based on regular participant count and platform
  -- Knowledge room: only myself as regular participant
  IF _regular_participant_count = 0 AND _participant_count = 1 AND
     EXISTS (SELECT 1 FROM room_participants WHERE room_id = _room_id AND contact_id = _myself_contact_id) THEN
    _room_type := 'knowledge';
  ELSIF _regular_participant_count = 2 THEN
    _room_type := 'person';
  ELSIF _regular_participant_count > 2 THEN
    _room_type := 'group';
  ELSE
    _room_type := 'unknown';
  END IF;

  -- Consolidated room_state update
  UPDATE room_state SET
    room_type = _room_type,
    room_platform = _room_platform,
    participant_count = _participant_count,
    last_message_text = CASE
      WHEN _msg_id IS NOT NULL THEN _message_body
      ELSE last_message_text
    END,
    last_message_id = CASE
      WHEN _msg_id IS NOT NULL THEN _msg_id
      ELSE last_message_id
    END,
    avatar = COALESCE(_room_avatar, avatar),
    updated_at = CASE
      WHEN display_name IS DISTINCT FROM _room_name
        OR avatar IS DISTINCT FROM _room_avatar
        OR room_platform IS DISTINCT FROM _room_platform
      THEN _created
      ELSE updated_at
    END
  WHERE room_id = _room_id;

END;
$$;


ALTER PROCEDURE public.process_message_body(IN input_json jsonb) OWNER TO postgres;

--
-- Name: process_message_type(uuid, text); Type: PROCEDURE; Schema: public; Owner: postgres
--

CREATE PROCEDURE public.process_message_type(IN new_message_id uuid, IN input_event_id text)
    LANGUAGE plpgsql
    AS $$
DECLARE
    input_json JSONB;
    input_date TIMESTAMP;
    recent_event BOOLEAN;
    is_audio BOOLEAN;
    is_bookmark_search BOOLEAN;
    is_alicia_chat BOOLEAN;
    new_conversation_search BOOLEAN;
    has_pdf BOOLEAN;
    message_record RECORD;
BEGIN
    -- Get the message details from the new messages table
    SELECT m.*, r.source_id AS room_source_id
    INTO message_record
    FROM messages m
    JOIN rooms r ON m.room_id = r.room_id
    WHERE m.event_id = input_event_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Message with event_id % not found', input_event_id;
    END IF;
    
    input_date := message_record.event_datetime;
    
    -- Get the raw message content if needed
    SELECT content INTO input_json 
    FROM raw_messages 
    WHERE external_id = input_event_id;
    
    -- Check if this is a recent event
    SELECT input_date > (NOW() - INTERVAL '1 MINUTE') INTO recent_event;
    
    -- Check message type
    SELECT message_record.msgtype = 'm.audio' INTO is_audio;
    SELECT message_record.msgtype = 'm.file' 
      AND message_record.body ILIKE '%.pdf'
      INTO has_pdf;
    
    -- Check specific room conditions
    SELECT message_record.room_source_id = '!gFlcwrWehFNHLTQzIV:diezel.org'
      AND message_record.sender_contact_id = '5a54e7ec-a4f4-42f8-881e-b0c41b910d97'
      INTO is_bookmark_search;
      
    SELECT message_record.room_source_id = '!oPVFvmbjLwybKAzBFL:diezel.org'
      AND message_record.sender_contact_id = '5a54e7ec-a4f4-42f8-881e-b0c41b910d97'
      INTO new_conversation_search;
      
    SELECT message_record.room_source_id = '!VRftofNawAybnACiJn:diezel.org'
      AND message_record.sender_contact_id = '5a54e7ec-a4f4-42f8-881e-b0c41b910d97'
      INTO is_alicia_chat;

    -- Send notifications based on message type
    IF is_audio THEN
        RAISE NOTICE 'Created audio message %', input_event_id;
        PERFORM pg_notify('new_audio_message', input_event_id);
    END IF;

    IF is_bookmark_search THEN
        PERFORM pg_notify('new_bookmark_search', input_event_id);
    END IF;
    
    IF new_conversation_search THEN
        PERFORM pg_notify('new_conversation_search', input_event_id);
    END IF;

    IF is_alicia_chat THEN
        PERFORM pg_notify('new_alicia_chat', input_event_id);
    END IF;
    
    IF has_pdf THEN
        PERFORM pg_notify('message_with_pdf', input_event_id);
    END IF;

    -- General notifications
    PERFORM pg_notify('new_message', input_event_id);
    
    IF recent_event THEN
        PERFORM pg_notify('new_message_recent', input_event_id);
    ELSE
        PERFORM pg_notify('new_message_backfill', input_event_id);
    END IF;
END;
$$;


ALTER PROCEDURE public.process_message_type(IN new_message_id uuid, IN input_event_id text) OWNER TO postgres;

--
-- Name: process_message_type_eval(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.process_message_type_eval() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    CALL process_message_type(NEW.message_id, NEW.event_id);
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.process_message_type_eval() OWNER TO postgres;

--
-- Name: process_new_message_into_session(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.process_new_message_into_session() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Call the stored procedure with the new message's ID and an 8-hour time_gap
    CALL add_message_to_session(NEW.message_id, INTERVAL '8 hours');

    -- Return NEW to indicate successful processing of the new row
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.process_new_message_into_session() OWNER TO postgres;

--
-- Name: process_raw_message_to_message(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.process_raw_message_to_message() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    CALL process_message_body(NEW."content"::jsonb);
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.process_raw_message_to_message() OWNER TO postgres;

--
-- Name: record_room_name(text, text, timestamp without time zone); Type: PROCEDURE; Schema: public; Owner: gardener
--

CREATE PROCEDURE public.record_room_name(IN _id text, IN name text, IN clock_time timestamp without time zone)
    LANGUAGE plpgsql
    AS $$
DECLARE
BEGIN
END;
$$;


ALTER PROCEDURE public.record_room_name(IN _id text, IN name text, IN clock_time timestamp without time zone) OWNER TO gardener;

--
-- Name: remove_contact_tag(uuid, text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.remove_contact_tag(contact_id uuid, tagname text) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    found_tag_id UUID;
BEGIN
    -- Get the tag ID
    SELECT tag_id INTO found_tag_id FROM contact_tagnames WHERE name = tagname;

    -- If the tag exists, delete the contact_tag relation
    IF FOUND THEN
        DELETE FROM contact_tags ct WHERE ct.contact_id = remove_contact_tag.contact_id AND ct.tag_id = found_tag_id;
    END IF;
END;
$$;


ALTER FUNCTION public.remove_contact_tag(contact_id uuid, tagname text) OWNER TO postgres;

--
-- Name: remove_tag(uuid, text); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.remove_tag(item_id uuid, tagname text) RETURNS void
    LANGUAGE plpgsql
    AS $$
DECLARE
    found_tag_id UUID;
BEGIN
    -- Get the tag ID
    SELECT id INTO found_tag_id FROM tags WHERE name = tagname;

    -- If the tag exists, delete the item_tag relation
    IF FOUND THEN
        DELETE FROM item_tags it WHERE it.item_id = remove_tag.item_id AND it.tag_id = found_tag_id;
    END IF;
END;
$$;


ALTER FUNCTION public.remove_tag(item_id uuid, tagname text) OWNER TO gardener;

--
-- Name: set_slug_from_name(); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.set_slug_from_name() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  NEW.slug := slugify(NEW.title);
  RETURN NEW;
END
$$;


ALTER FUNCTION public.set_slug_from_name() OWNER TO gardener;

--
-- Name: slugify(text); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.slugify(value text) RETURNS text
    LANGUAGE sql IMMUTABLE STRICT
    AS $_$
  -- removes accents (diacritic signs) from a given string --
  WITH "unaccented" AS (
    SELECT unaccent("value") AS "value"
  ),
  -- lowercases the string
  "lowercase" AS (
    SELECT lower("value") AS "value"
    FROM "unaccented"
  ),
  -- replaces anything that's not a letter, number, hyphen('-'), or underscore('_') with a hyphen('-')
  "hyphenated" AS (
    SELECT regexp_replace("value", '[^a-z0-9\\-_]+', '-', 'gi') AS "value"
    FROM "lowercase"
  ),
  -- trims hyphens('-') if they exist on the head or tail of the string
  "trimmed" AS (
    SELECT regexp_replace(regexp_replace("value", '\\-+$', ''), '^\\-', '') AS "value"
    FROM "hyphenated"
  )
  SELECT "value" FROM "trimmed";
$_$;


ALTER FUNCTION public.slugify(value text) OWNER TO gardener;

--
-- Name: strip_html_tags(text); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.strip_html_tags(text_with_html text) RETURNS text
    LANGUAGE plpgsql
    AS $$
DECLARE
  result TEXT;
BEGIN
  -- Simple HTML tag removal (this is a basic implementation)
  -- For a more robust solution, consider using a proper HTML parser
  result := regexp_replace(text_with_html, '<[^>]*>', '', 'g');
  -- Replace common HTML entities
  result := regexp_replace(result, '&nbsp;', ' ', 'g');
  result := regexp_replace(result, '&amp;', '&', 'g');
  result := regexp_replace(result, '&lt;', '<', 'g');
  result := regexp_replace(result, '&gt;', '>', 'g');
  result := regexp_replace(result, '&quot;', '"', 'g');
  result := regexp_replace(result, '&#39;', '''', 'g');
  -- Trim whitespace
  result := trim(result);
  RETURN result;
END;
$$;


ALTER FUNCTION public.strip_html_tags(text_with_html text) OWNER TO postgres;

--
-- Name: update_category_name(uuid, text); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.update_category_name(category_id_param uuid, new_name text) RETURNS void
    LANGUAGE plpgsql
    AS $$
BEGIN
    UPDATE categories
    SET name = new_name
    WHERE categories.category_id = category_id_param;
END;
$$;


ALTER FUNCTION public.update_category_name(category_id_param uuid, new_name text) OWNER TO gardener;

--
-- Name: update_modified_column(); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.update_modified_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.modified = EXTRACT(EPOCH FROM CURRENT_TIMESTAMP);
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_modified_column() OWNER TO gardener;

--
-- Name: update_tag_last_activity(); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.update_tag_last_activity() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    UPDATE tags
    SET last_activity = EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)
    WHERE id = NEW.tag_id OR id = OLD.tag_id;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_tag_last_activity() OWNER TO gardener;

--
-- Name: update_tag_modified_column(); Type: FUNCTION; Schema: public; Owner: gardener
--

CREATE FUNCTION public.update_tag_modified_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.modified = EXTRACT(EPOCH FROM CURRENT_TIMESTAMP);
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_tag_modified_column() OWNER TO gardener;

--
-- Name: view_session_details(uuid); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.view_session_details(session_id_param uuid) RETURNS TABLE(message_id uuid, event_id text, sender_name text, sender_contact_id uuid, room_id uuid, room_name text, message_type text, message_classification text, body text, formatted_body text, event_datetime timestamp without time zone, is_edited boolean, is_reply boolean, reply_to_event_id text)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT 
        m.message_id,
        m.event_id,
        c.name AS sender_name,
        m.sender_contact_id,
        m.room_id,
        r.display_name AS room_name,
        m.message_type,
        m.message_classification,
        m.body,
        m.formatted_body,
        m.event_datetime,
        m.is_edited,
        m.is_reply,
        m.reply_to_event_id
    FROM 
        messages m
    JOIN 
        session_message sm ON m.message_id = sm.message_id
    JOIN 
        contacts c ON m.sender_contact_id = c.contact_id
    JOIN 
        rooms r ON m.room_id = r.room_id
    WHERE 
        sm.session_id = session_id_param
    ORDER BY 
        m.event_datetime;
END;
$$;


ALTER FUNCTION public.view_session_details(session_id_param uuid) OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: alicia_conversations; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.alicia_conversations (
    id text DEFAULT public.generate_random_id('ac'::text) NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.alicia_conversations OWNER TO gardener;

--
-- Name: alicia_message; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.alicia_message (
    id text DEFAULT public.generate_random_id('am'::text) NOT NULL,
    role text NOT NULL,
    contents text NOT NULL,
    previous_id text,
    conversation_id text,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.alicia_message OWNER TO gardener;

--
-- Name: alicia_meta; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.alicia_meta (
    id text DEFAULT public.generate_random_id('at'::text) NOT NULL,
    ref text,
    contents jsonb NOT NULL,
    conversation_id text,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.alicia_meta OWNER TO gardener;

--
-- Name: bookmark_category; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.bookmark_category (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    bookmark_id uuid,
    category_id uuid
);


ALTER TABLE public.bookmark_category OWNER TO gardener;

--
-- Name: bookmark_content_references; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.bookmark_content_references (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    bookmark_id uuid,
    content text,
    strategy text,
    embedding public.vector(1024),
    created_at timestamp without time zone DEFAULT now(),
    extra jsonb DEFAULT '{}'::jsonb
);

ALTER TABLE ONLY public.bookmark_content_references REPLICA IDENTITY FULL;


ALTER TABLE public.bookmark_content_references OWNER TO gardener;

--
-- Name: bookmark_evaluations; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.bookmark_evaluations (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    bookmark_id uuid NOT NULL,
    failed_fetch boolean,
    summary_id text,
    summary_comments text,
    summary_eval integer,
    questions_eval integer,
    questions_ids jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.bookmark_evaluations OWNER TO gardener;

--
-- Name: bookmark_sources; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.bookmark_sources (
    source_id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    bookmark_id uuid,
    source_uri text,
    raw_source json
);


ALTER TABLE public.bookmark_sources OWNER TO gardener;

--
-- Name: bookmark_titles; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.bookmark_titles (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    bookmark_id uuid,
    title text,
    source text
);


ALTER TABLE public.bookmark_titles OWNER TO gardener;

--
-- Name: bookmarks; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.bookmarks (
    bookmark_id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    url text NOT NULL,
    creation_date timestamp without time zone NOT NULL
);


ALTER TABLE public.bookmarks OWNER TO gardener;

--
-- Name: browser_history; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.browser_history (
    id integer NOT NULL,
    url text NOT NULL,
    title text,
    visit_date timestamp without time zone NOT NULL,
    typed boolean,
    hidden boolean,
    imported_from_firefox_place_id integer,
    imported_from_firefox_visit_id integer,
    domain text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.browser_history OWNER TO gardener;

--
-- Name: browser_history_id_seq; Type: SEQUENCE; Schema: public; Owner: gardener
--

CREATE SEQUENCE public.browser_history_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.browser_history_id_seq OWNER TO gardener;

--
-- Name: browser_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: gardener
--

ALTER SEQUENCE public.browser_history_id_seq OWNED BY public.browser_history.id;


--
-- Name: categories; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.categories (
    category_id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL
);


ALTER TABLE public.categories OWNER TO gardener;

--
-- Name: category_sources; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.category_sources (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    category_id uuid,
    source_uri text,
    raw_source json
);


ALTER TABLE public.category_sources OWNER TO gardener;

--
-- Name: configurations; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.configurations (
    config_id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    key text NOT NULL,
    value text NOT NULL,
    is_secret boolean DEFAULT false NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.configurations OWNER TO gardener;

--
-- Name: contact_evals; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.contact_evals (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    contact_id uuid NOT NULL,
    importance integer,
    closeness integer,
    fondness integer,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    CONSTRAINT contact_evals_closeness_check CHECK (((closeness >= 1) AND (closeness <= 5))),
    CONSTRAINT contact_evals_fondness_check CHECK (((fondness >= 1) AND (fondness <= 5))),
    CONSTRAINT contact_evals_importance_check CHECK (((importance >= 1) AND (importance <= 5)))
);

ALTER TABLE ONLY public.contact_evals REPLICA IDENTITY FULL;


ALTER TABLE public.contact_evals OWNER TO gardener;

--
-- Name: contact_known_avatars; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.contact_known_avatars (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    contact_id uuid NOT NULL,
    avatar text NOT NULL,
    earliest_date timestamp without time zone
);

ALTER TABLE ONLY public.contact_known_avatars REPLICA IDENTITY FULL;


ALTER TABLE public.contact_known_avatars OWNER TO gardener;

--
-- Name: contact_known_names; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.contact_known_names (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    contact_id uuid NOT NULL,
    name text NOT NULL
);

ALTER TABLE ONLY public.contact_known_names REPLICA IDENTITY FULL;


ALTER TABLE public.contact_known_names OWNER TO gardener;

--
-- Name: contact_sources; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.contact_sources (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    contact_id uuid,
    source_id text NOT NULL,
    source_name text NOT NULL
);


ALTER TABLE public.contact_sources OWNER TO gardener;

--
-- Name: contact_stats; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.contact_stats (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    contact_id uuid NOT NULL,
    last_week_messages integer,
    groups_in_common integer
);

ALTER TABLE ONLY public.contact_stats REPLICA IDENTITY FULL;


ALTER TABLE public.contact_stats OWNER TO gardener;

--
-- Name: contact_tagnames; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.contact_tagnames (
    tag_id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);

ALTER TABLE ONLY public.contact_tagnames REPLICA IDENTITY FULL;


ALTER TABLE public.contact_tagnames OWNER TO gardener;

--
-- Name: contact_tags; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.contact_tags (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    contact_id uuid NOT NULL,
    tag_id uuid NOT NULL
);

ALTER TABLE ONLY public.contact_tags REPLICA IDENTITY FULL;


ALTER TABLE public.contact_tags OWNER TO gardener;

--
-- Name: contacts; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.contacts (
    contact_id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    email text,
    phone text,
    creation_date timestamp without time zone DEFAULT now() NOT NULL,
    last_update timestamp without time zone DEFAULT now() NOT NULL,
    birthday text,
    notes text,
    extras jsonb DEFAULT '{}'::jsonb
);

ALTER TABLE ONLY public.contacts REPLICA IDENTITY FULL;


ALTER TABLE public.contacts OWNER TO gardener;

--
-- Name: dispatch_analysis; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.dispatch_analysis (
    id integer NOT NULL,
    transcription_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    is_edited boolean DEFAULT false NOT NULL,
    parent_id integer,
    prompt_used text NOT NULL,
    raw_response text NOT NULL,
    extra jsonb DEFAULT '{}'::jsonb NOT NULL
);


ALTER TABLE public.dispatch_analysis OWNER TO gardener;

--
-- Name: dispatch_analysis_id_seq; Type: SEQUENCE; Schema: public; Owner: gardener
--

CREATE SEQUENCE public.dispatch_analysis_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dispatch_analysis_id_seq OWNER TO gardener;

--
-- Name: dispatch_analysis_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: gardener
--

ALTER SEQUENCE public.dispatch_analysis_id_seq OWNED BY public.dispatch_analysis.id;


--
-- Name: dispatch_audio_raw; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.dispatch_audio_raw (
    id integer NOT NULL,
    file_data bytea NOT NULL,
    file_name character varying NOT NULL,
    file_type character varying NOT NULL,
    file_size integer NOT NULL,
    duration real,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL
);


ALTER TABLE public.dispatch_audio_raw OWNER TO gardener;

--
-- Name: dispatch_audio_raw_id_seq; Type: SEQUENCE; Schema: public; Owner: gardener
--

CREATE SEQUENCE public.dispatch_audio_raw_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dispatch_audio_raw_id_seq OWNER TO gardener;

--
-- Name: dispatch_audio_raw_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: gardener
--

ALTER SEQUENCE public.dispatch_audio_raw_id_seq OWNED BY public.dispatch_audio_raw.id;


--
-- Name: dispatch_bulletpoint; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.dispatch_bulletpoint (
    id integer NOT NULL,
    analysis_id integer NOT NULL,
    from_position integer NOT NULL,
    to_position integer NOT NULL,
    info text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    order_index integer NOT NULL
);


ALTER TABLE public.dispatch_bulletpoint OWNER TO gardener;

--
-- Name: dispatch_bulletpoint_id_seq; Type: SEQUENCE; Schema: public; Owner: gardener
--

CREATE SEQUENCE public.dispatch_bulletpoint_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dispatch_bulletpoint_id_seq OWNER TO gardener;

--
-- Name: dispatch_bulletpoint_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: gardener
--

ALTER SEQUENCE public.dispatch_bulletpoint_id_seq OWNED BY public.dispatch_bulletpoint.id;


--
-- Name: dispatch_entity_merge; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.dispatch_entity_merge (
    id integer NOT NULL,
    entity_id integer NOT NULL,
    merged_entity_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_by character varying,
    notes text
);


ALTER TABLE public.dispatch_entity_merge OWNER TO gardener;

--
-- Name: dispatch_entity_merge_id_seq; Type: SEQUENCE; Schema: public; Owner: gardener
--

CREATE SEQUENCE public.dispatch_entity_merge_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dispatch_entity_merge_id_seq OWNER TO gardener;

--
-- Name: dispatch_entity_merge_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: gardener
--

ALTER SEQUENCE public.dispatch_entity_merge_id_seq OWNED BY public.dispatch_entity_merge.id;


--
-- Name: dispatch_extracted_entity; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.dispatch_extracted_entity (
    id integer NOT NULL,
    bulletpoint_id integer NOT NULL,
    entity_text character varying NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.dispatch_extracted_entity OWNER TO gardener;

--
-- Name: dispatch_extracted_entity_id_seq; Type: SEQUENCE; Schema: public; Owner: gardener
--

CREATE SEQUENCE public.dispatch_extracted_entity_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dispatch_extracted_entity_id_seq OWNER TO gardener;

--
-- Name: dispatch_extracted_entity_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: gardener
--

ALTER SEQUENCE public.dispatch_extracted_entity_id_seq OWNED BY public.dispatch_extracted_entity.id;


--
-- Name: dispatch_job; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.dispatch_job (
    id integer NOT NULL,
    title character varying,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    audio_id integer,
    transcription_id integer,
    analysis_id integer,
    status character varying DEFAULT 'pending'::character varying NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    CONSTRAINT dispatch_job_status_check CHECK (((status)::text = ANY (ARRAY[('pending'::character varying)::text, ('in_progress'::character varying)::text, ('completed'::character varying)::text, ('failed'::character varying)::text])))
);


ALTER TABLE public.dispatch_job OWNER TO gardener;

--
-- Name: dispatch_job_id_seq; Type: SEQUENCE; Schema: public; Owner: gardener
--

CREATE SEQUENCE public.dispatch_job_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dispatch_job_id_seq OWNER TO gardener;

--
-- Name: dispatch_job_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: gardener
--

ALTER SEQUENCE public.dispatch_job_id_seq OWNED BY public.dispatch_job.id;


--
-- Name: dispatch_transcription; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.dispatch_transcription (
    id integer NOT NULL,
    audio_id integer NOT NULL,
    transcription_text text NOT NULL,
    is_edited boolean DEFAULT false NOT NULL,
    parent_id integer,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    extra jsonb DEFAULT '{}'::jsonb NOT NULL
);


ALTER TABLE public.dispatch_transcription OWNER TO gardener;

--
-- Name: dispatch_transcription_id_seq; Type: SEQUENCE; Schema: public; Owner: gardener
--

CREATE SEQUENCE public.dispatch_transcription_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.dispatch_transcription_id_seq OWNER TO gardener;

--
-- Name: dispatch_transcription_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: gardener
--

ALTER SEQUENCE public.dispatch_transcription_id_seq OWNED BY public.dispatch_transcription.id;


--
-- Name: entities; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.entities (
    entity_id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    description text,
    properties jsonb DEFAULT '{}'::jsonb,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at timestamp without time zone
);


ALTER TABLE public.entities OWNER TO gardener;

--
-- Name: entity_references; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.entity_references (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    source_type text NOT NULL,
    source_id uuid NOT NULL,
    entity_id uuid NOT NULL,
    reference_text text NOT NULL,
    "position" integer,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.entity_references OWNER TO gardener;

--
-- Name: entity_relationships; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.entity_relationships (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    entity_id uuid NOT NULL,
    related_type text NOT NULL,
    related_id uuid NOT NULL,
    relationship_type text NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.entity_relationships OWNER TO gardener;

--
-- Name: item_tags; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.item_tags (
    item_id uuid NOT NULL,
    tag_id uuid NOT NULL
);


ALTER TABLE public.item_tags OWNER TO gardener;

--
-- Name: items; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.items (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    title text,
    slug text,
    contents text,
    created bigint DEFAULT EXTRACT(epoch FROM CURRENT_TIMESTAMP),
    modified bigint DEFAULT EXTRACT(epoch FROM CURRENT_TIMESTAMP)
);


ALTER TABLE public.items OWNER TO gardener;

--
-- Name: tags; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.tags (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    created bigint DEFAULT EXTRACT(epoch FROM CURRENT_TIMESTAMP),
    modified bigint DEFAULT EXTRACT(epoch FROM CURRENT_TIMESTAMP),
    last_activity bigint
);


ALTER TABLE public.tags OWNER TO gardener;

--
-- Name: full_items; Type: VIEW; Schema: public; Owner: gardener
--

CREATE VIEW public.full_items AS
 SELECT i.id,
    i.title,
    i.contents,
    i.created,
    i.modified,
    string_agg(t.name, ','::text ORDER BY t.name) AS tags
   FROM ((public.items i
     LEFT JOIN public.item_tags it ON ((i.id = it.item_id)))
     LEFT JOIN public.tags t ON ((it.tag_id = t.id)))
  GROUP BY i.id, i.title, i.contents, i.created, i.modified;


ALTER VIEW public.full_items OWNER TO gardener;

--
-- Name: http_responses; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.http_responses (
    response_id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    bookmark_id uuid,
    status_code integer,
    headers text,
    content bytea,
    fetch_date timestamp without time zone
);


ALTER TABLE public.http_responses OWNER TO gardener;

--
-- Name: item_semantic_index; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.item_semantic_index (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    item_id uuid,
    embedding public.vector(1024)
);

ALTER TABLE ONLY public.item_semantic_index REPLICA IDENTITY FULL;


ALTER TABLE public.item_semantic_index OWNER TO gardener;

--
-- Name: message_text_representation; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.message_text_representation (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    message_id uuid NOT NULL,
    text_content text NOT NULL,
    source_type text NOT NULL,
    search_vector tsvector,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.message_text_representation OWNER TO gardener;

--
-- Name: messages_old; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.messages_old (
    message_id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    sender_id uuid NOT NULL,
    room_id uuid NOT NULL,
    content jsonb,
    external_id text NOT NULL,
    created_at timestamp without time zone
);


ALTER TABLE public.messages_old OWNER TO gardener;

--
-- Name: rooms; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.rooms (
    room_id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    source_id text NOT NULL,
    display_name text,
    user_defined_name text,
    last_activity timestamp without time zone
);

ALTER TABLE ONLY public.rooms REPLICA IDENTITY FULL;


ALTER TABLE public.rooms OWNER TO gardener;

--
-- Name: message_view; Type: VIEW; Schema: public; Owner: gardener
--

CREATE VIEW public.message_view AS
 SELECT m.message_id,
    r.room_id,
    r.display_name AS room,
    sc.contact_id AS from_id,
    sc.name AS "from",
    m.created_at AS "when",
    (m.content ->> 'body'::text) AS content
   FROM ((public.messages_old m
     JOIN public.rooms r USING (room_id))
     LEFT JOIN public.contacts sc ON ((m.sender_id = sc.contact_id)))
  WHERE ((m.content ->> 'msgtype'::text) = 'm.text'::text)
  ORDER BY m.created_at DESC;


ALTER VIEW public.message_view OWNER TO gardener;

--
-- Name: messages; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.messages (
    message_id uuid DEFAULT gen_random_uuid() NOT NULL,
    event_id text NOT NULL,
    event_datetime timestamp without time zone,
    origin_server_ts bigint,
    sender_contact_id uuid NOT NULL,
    room_id uuid NOT NULL,
    message_type text,
    message_classification text,
    body text,
    formatted_body text,
    format text,
    msgtype text,
    is_edited boolean DEFAULT false,
    is_reply boolean DEFAULT false,
    reply_to_event_id text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY public.messages REPLICA IDENTITY FULL;


ALTER TABLE public.messages OWNER TO gardener;

--
-- Name: messages_edit_history; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.messages_edit_history (
    edit_id uuid DEFAULT gen_random_uuid() NOT NULL,
    message_id uuid NOT NULL,
    previous_body text,
    previous_formatted_body text,
    edit_timestamp timestamp without time zone NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.messages_edit_history OWNER TO gardener;

--
-- Name: messages_media; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.messages_media (
    media_id uuid DEFAULT gen_random_uuid() NOT NULL,
    message_id uuid NOT NULL,
    url text,
    mimetype text,
    size integer,
    width integer,
    height integer,
    duration integer,
    filename text,
    is_encrypted boolean DEFAULT false,
    thumbnail_url text,
    geo_uri text,
    location_description text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY public.messages_media REPLICA IDENTITY FULL;


ALTER TABLE public.messages_media OWNER TO gardener;

--
-- Name: messages_mentions; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.messages_mentions (
    mention_id uuid DEFAULT gen_random_uuid() NOT NULL,
    message_id uuid NOT NULL,
    contact_id uuid,
    room_mention boolean DEFAULT false,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY public.messages_mentions REPLICA IDENTITY FULL;


ALTER TABLE public.messages_mentions OWNER TO gardener;

--
-- Name: messages_reactions; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.messages_reactions (
    reaction_id uuid DEFAULT gen_random_uuid() NOT NULL,
    message_id uuid NOT NULL,
    target_event_id text NOT NULL,
    sender_contact_id uuid NOT NULL,
    key text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY public.messages_reactions REPLICA IDENTITY FULL;


ALTER TABLE public.messages_reactions OWNER TO gardener;

--
-- Name: messages_relations; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.messages_relations (
    relation_id uuid DEFAULT gen_random_uuid() NOT NULL,
    source_message_id uuid NOT NULL,
    target_event_id text NOT NULL,
    relation_type text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY public.messages_relations REPLICA IDENTITY FULL;


ALTER TABLE public.messages_relations OWNER TO gardener;

--
-- Name: observations; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.observations (
    observation_id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    data jsonb NOT NULL,
    type text,
    source text,
    tags text,
    parent uuid,
    ref uuid,
    creation_date bigint DEFAULT EXTRACT(epoch FROM CURRENT_TIMESTAMP) NOT NULL
);


ALTER TABLE public.observations OWNER TO gardener;

--
-- Name: processed_contents; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.processed_contents (
    processed_content_id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    bookmark_id uuid,
    strategy_used text,
    processed_content text
);


ALTER TABLE public.processed_contents OWNER TO gardener;

--
-- Name: raw_messages; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.raw_messages (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    external_id text NOT NULL,
    content json,
    created_at timestamp without time zone
);


ALTER TABLE public.raw_messages OWNER TO gardener;

--
-- Name: room_known_avatars; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.room_known_avatars (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    room_id uuid,
    avatar text,
    earliest_date timestamp without time zone,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.room_known_avatars OWNER TO gardener;

--
-- Name: room_known_names; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.room_known_names (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    room_id uuid NOT NULL,
    name text NOT NULL,
    last_time timestamp without time zone
);


ALTER TABLE public.room_known_names OWNER TO gardener;

--
-- Name: room_participants; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.room_participants (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    room_id uuid NOT NULL,
    contact_id uuid NOT NULL,
    known_last_presence timestamp without time zone,
    known_last_exit timestamp without time zone
);

ALTER TABLE ONLY public.room_participants REPLICA IDENTITY FULL;


ALTER TABLE public.room_participants OWNER TO gardener;

--
-- Name: room_state; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.room_state (
    room_id uuid NOT NULL,
    room_type text,
    room_platform text,
    display_name text,
    subtitle text,
    avatar text,
    participant_count integer DEFAULT 0,
    unread_counts integer DEFAULT 0,
    unread_highlights integer DEFAULT 0,
    last_activity timestamp without time zone,
    last_message_id uuid,
    last_message_text text,
    is_muted boolean DEFAULT false,
    is_hidden boolean DEFAULT false,
    is_favorite boolean DEFAULT false,
    extra jsonb,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.room_state OWNER TO gardener;

--
-- Name: session_message; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.session_message (
    session_id uuid NOT NULL,
    message_id uuid NOT NULL
);


ALTER TABLE public.session_message OWNER TO gardener;

--
-- Name: session_summaries; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.session_summaries (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    session_id uuid NOT NULL,
    summary text,
    embedding public.vector(1024),
    strategy text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.session_summaries OWNER TO gardener;

--
-- Name: sessions; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.sessions (
    session_id uuid DEFAULT gen_random_uuid() NOT NULL,
    room_id uuid NOT NULL,
    first_date_time timestamp without time zone NOT NULL,
    first_message_id uuid,
    last_message_id uuid,
    last_date_time timestamp without time zone,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.sessions OWNER TO gardener;

--
-- Name: social_posts; Type: TABLE; Schema: public; Owner: gardener
--

CREATE TABLE public.social_posts (
    post_id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    content text NOT NULL,
    twitter_post_id text,
    bluesky_post_id text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp with time zone,
    status text DEFAULT 'pending'::text NOT NULL,
    error_message text
);


ALTER TABLE public.social_posts OWNER TO gardener;

--
-- Name: browser_history id; Type: DEFAULT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.browser_history ALTER COLUMN id SET DEFAULT nextval('public.browser_history_id_seq'::regclass);


--
-- Name: dispatch_analysis id; Type: DEFAULT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_analysis ALTER COLUMN id SET DEFAULT nextval('public.dispatch_analysis_id_seq'::regclass);


--
-- Name: dispatch_audio_raw id; Type: DEFAULT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_audio_raw ALTER COLUMN id SET DEFAULT nextval('public.dispatch_audio_raw_id_seq'::regclass);


--
-- Name: dispatch_bulletpoint id; Type: DEFAULT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_bulletpoint ALTER COLUMN id SET DEFAULT nextval('public.dispatch_bulletpoint_id_seq'::regclass);


--
-- Name: dispatch_entity_merge id; Type: DEFAULT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_entity_merge ALTER COLUMN id SET DEFAULT nextval('public.dispatch_entity_merge_id_seq'::regclass);


--
-- Name: dispatch_extracted_entity id; Type: DEFAULT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_extracted_entity ALTER COLUMN id SET DEFAULT nextval('public.dispatch_extracted_entity_id_seq'::regclass);


--
-- Name: dispatch_job id; Type: DEFAULT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_job ALTER COLUMN id SET DEFAULT nextval('public.dispatch_job_id_seq'::regclass);


--
-- Name: dispatch_transcription id; Type: DEFAULT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_transcription ALTER COLUMN id SET DEFAULT nextval('public.dispatch_transcription_id_seq'::regclass);


--
-- Name: alicia_conversations alicia_conversations_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.alicia_conversations
    ADD CONSTRAINT alicia_conversations_pkey PRIMARY KEY (id);


--
-- Name: alicia_message alicia_message_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.alicia_message
    ADD CONSTRAINT alicia_message_pkey PRIMARY KEY (id);


--
-- Name: alicia_meta alicia_meta_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.alicia_meta
    ADD CONSTRAINT alicia_meta_pkey PRIMARY KEY (id);


--
-- Name: bookmark_category bookmark_category_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.bookmark_category
    ADD CONSTRAINT bookmark_category_pkey PRIMARY KEY (id);


--
-- Name: bookmark_evaluations bookmark_evaluations_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.bookmark_evaluations
    ADD CONSTRAINT bookmark_evaluations_pkey PRIMARY KEY (id);


--
-- Name: bookmark_sources bookmark_sources_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.bookmark_sources
    ADD CONSTRAINT bookmark_sources_pkey PRIMARY KEY (source_id);


--
-- Name: bookmark_titles bookmark_titles_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.bookmark_titles
    ADD CONSTRAINT bookmark_titles_pkey PRIMARY KEY (id);


--
-- Name: bookmarks bookmarks_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.bookmarks
    ADD CONSTRAINT bookmarks_pkey PRIMARY KEY (bookmark_id);


--
-- Name: browser_history browser_history_imported_from_firefox_visit_id_key; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.browser_history
    ADD CONSTRAINT browser_history_imported_from_firefox_visit_id_key UNIQUE (imported_from_firefox_visit_id);


--
-- Name: browser_history browser_history_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.browser_history
    ADD CONSTRAINT browser_history_pkey PRIMARY KEY (id);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (category_id);


--
-- Name: category_sources category_sources_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.category_sources
    ADD CONSTRAINT category_sources_pkey PRIMARY KEY (id);


--
-- Name: configurations configurations_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.configurations
    ADD CONSTRAINT configurations_pkey PRIMARY KEY (config_id);


--
-- Name: contact_evals contact_evals_contact_id_key; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contact_evals
    ADD CONSTRAINT contact_evals_contact_id_key UNIQUE (contact_id);


--
-- Name: contact_evals contact_evals_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contact_evals
    ADD CONSTRAINT contact_evals_pkey PRIMARY KEY (id);


--
-- Name: contact_known_avatars contact_known_avatars_contact_id_avatar_key; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contact_known_avatars
    ADD CONSTRAINT contact_known_avatars_contact_id_avatar_key UNIQUE (contact_id, avatar);


--
-- Name: contact_known_avatars contact_known_avatars_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contact_known_avatars
    ADD CONSTRAINT contact_known_avatars_pkey PRIMARY KEY (id);


--
-- Name: contact_known_names contact_known_names_contact_id_name_key; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contact_known_names
    ADD CONSTRAINT contact_known_names_contact_id_name_key UNIQUE (contact_id, name);


--
-- Name: contact_known_names contact_known_names_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contact_known_names
    ADD CONSTRAINT contact_known_names_pkey PRIMARY KEY (id);


--
-- Name: contact_sources contact_sources_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contact_sources
    ADD CONSTRAINT contact_sources_pkey PRIMARY KEY (id);


--
-- Name: contact_stats contact_stats_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contact_stats
    ADD CONSTRAINT contact_stats_pkey PRIMARY KEY (id);


--
-- Name: contact_tagnames contact_tagnames_name_key; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contact_tagnames
    ADD CONSTRAINT contact_tagnames_name_key UNIQUE (name);


--
-- Name: contact_tagnames contact_tagnames_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contact_tagnames
    ADD CONSTRAINT contact_tagnames_pkey PRIMARY KEY (tag_id);


--
-- Name: contact_tags contact_tags_contact_id_tag_id_key; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contact_tags
    ADD CONSTRAINT contact_tags_contact_id_tag_id_key UNIQUE (contact_id, tag_id);


--
-- Name: contact_tags contact_tags_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contact_tags
    ADD CONSTRAINT contact_tags_pkey PRIMARY KEY (id);


--
-- Name: contacts contacts_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contacts
    ADD CONSTRAINT contacts_pkey PRIMARY KEY (contact_id);


--
-- Name: dispatch_analysis dispatch_analysis_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_analysis
    ADD CONSTRAINT dispatch_analysis_pkey PRIMARY KEY (id);


--
-- Name: dispatch_audio_raw dispatch_audio_raw_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_audio_raw
    ADD CONSTRAINT dispatch_audio_raw_pkey PRIMARY KEY (id);


--
-- Name: dispatch_bulletpoint dispatch_bulletpoint_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_bulletpoint
    ADD CONSTRAINT dispatch_bulletpoint_pkey PRIMARY KEY (id);


--
-- Name: dispatch_entity_merge dispatch_entity_merge_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_entity_merge
    ADD CONSTRAINT dispatch_entity_merge_pkey PRIMARY KEY (id);


--
-- Name: dispatch_extracted_entity dispatch_extracted_entity_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_extracted_entity
    ADD CONSTRAINT dispatch_extracted_entity_pkey PRIMARY KEY (id);


--
-- Name: dispatch_job dispatch_job_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_job
    ADD CONSTRAINT dispatch_job_pkey PRIMARY KEY (id);


--
-- Name: dispatch_transcription dispatch_transcription_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_transcription
    ADD CONSTRAINT dispatch_transcription_pkey PRIMARY KEY (id);


--
-- Name: entities entities_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.entities
    ADD CONSTRAINT entities_pkey PRIMARY KEY (entity_id);


--
-- Name: entity_references entity_references_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.entity_references
    ADD CONSTRAINT entity_references_pkey PRIMARY KEY (id);


--
-- Name: entity_relationships entity_relationships_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.entity_relationships
    ADD CONSTRAINT entity_relationships_pkey PRIMARY KEY (id);


--
-- Name: http_responses http_responses_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.http_responses
    ADD CONSTRAINT http_responses_pkey PRIMARY KEY (response_id);


--
-- Name: item_tags item_tags_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.item_tags
    ADD CONSTRAINT item_tags_pkey PRIMARY KEY (item_id, tag_id);


--
-- Name: items items_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.items
    ADD CONSTRAINT items_pkey PRIMARY KEY (id);


--
-- Name: message_text_representation message_text_representation_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.message_text_representation
    ADD CONSTRAINT message_text_representation_pkey PRIMARY KEY (id);


--
-- Name: messages_edit_history messages_edit_history_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages_edit_history
    ADD CONSTRAINT messages_edit_history_pkey PRIMARY KEY (edit_id);


--
-- Name: messages messages_event_id_key; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_event_id_key UNIQUE (event_id);


--
-- Name: messages_old messages_external_id_key; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages_old
    ADD CONSTRAINT messages_external_id_key UNIQUE (external_id);


--
-- Name: messages_media messages_media_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages_media
    ADD CONSTRAINT messages_media_pkey PRIMARY KEY (media_id);


--
-- Name: messages_mentions messages_mentions_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages_mentions
    ADD CONSTRAINT messages_mentions_pkey PRIMARY KEY (mention_id);


--
-- Name: messages_old messages_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages_old
    ADD CONSTRAINT messages_pkey PRIMARY KEY (message_id);


--
-- Name: messages messages_pkey1; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pkey1 PRIMARY KEY (message_id);


--
-- Name: messages_reactions messages_reactions_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages_reactions
    ADD CONSTRAINT messages_reactions_pkey PRIMARY KEY (reaction_id);


--
-- Name: messages_relations messages_relations_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages_relations
    ADD CONSTRAINT messages_relations_pkey PRIMARY KEY (relation_id);


--
-- Name: observations observations_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.observations
    ADD CONSTRAINT observations_pkey PRIMARY KEY (observation_id);


--
-- Name: processed_contents processed_contents_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.processed_contents
    ADD CONSTRAINT processed_contents_pkey PRIMARY KEY (processed_content_id);


--
-- Name: raw_messages raw_messages_external_id_key; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.raw_messages
    ADD CONSTRAINT raw_messages_external_id_key UNIQUE (external_id);


--
-- Name: raw_messages raw_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.raw_messages
    ADD CONSTRAINT raw_messages_pkey PRIMARY KEY (id);


--
-- Name: room_known_avatars room_known_avatars_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.room_known_avatars
    ADD CONSTRAINT room_known_avatars_pkey PRIMARY KEY (id);


--
-- Name: room_known_avatars room_known_avatars_room_avatar_key; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.room_known_avatars
    ADD CONSTRAINT room_known_avatars_room_avatar_key UNIQUE (room_id, avatar);


--
-- Name: room_known_names room_known_names_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.room_known_names
    ADD CONSTRAINT room_known_names_pkey PRIMARY KEY (id);


--
-- Name: room_known_names room_known_names_room_id_name_key; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.room_known_names
    ADD CONSTRAINT room_known_names_room_id_name_key UNIQUE (room_id, name);


--
-- Name: room_participants room_participants_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.room_participants
    ADD CONSTRAINT room_participants_pkey PRIMARY KEY (id);


--
-- Name: room_participants room_participants_room_id_contact_id_key; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.room_participants
    ADD CONSTRAINT room_participants_room_id_contact_id_key UNIQUE (room_id, contact_id);


--
-- Name: room_state room_state_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.room_state
    ADD CONSTRAINT room_state_pkey PRIMARY KEY (room_id);


--
-- Name: rooms rooms_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.rooms
    ADD CONSTRAINT rooms_pkey PRIMARY KEY (room_id);


--
-- Name: rooms rooms_source_id_key; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.rooms
    ADD CONSTRAINT rooms_source_id_key UNIQUE (source_id);


--
-- Name: session_message session_message_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.session_message
    ADD CONSTRAINT session_message_pkey PRIMARY KEY (session_id, message_id);


--
-- Name: session_summaries session_summaries_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.session_summaries
    ADD CONSTRAINT session_summaries_pkey PRIMARY KEY (id);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (session_id);


--
-- Name: social_posts social_posts_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.social_posts
    ADD CONSTRAINT social_posts_pkey PRIMARY KEY (post_id);


--
-- Name: tags tags_name_key; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_name_key UNIQUE (name);


--
-- Name: tags tags_pkey; Type: CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_pkey PRIMARY KEY (id);


--
-- Name: bookmark_evaluations_bookmark_id_idx; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX bookmark_evaluations_bookmark_id_idx ON public.bookmark_evaluations USING btree (bookmark_id);


--
-- Name: idx_browser_history_domain; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_browser_history_domain ON public.browser_history USING btree (domain);


--
-- Name: idx_browser_history_visit_date; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_browser_history_visit_date ON public.browser_history USING btree (visit_date);


--
-- Name: idx_configurations_is_secret; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_configurations_is_secret ON public.configurations USING btree (is_secret);


--
-- Name: idx_configurations_key; Type: INDEX; Schema: public; Owner: gardener
--

CREATE UNIQUE INDEX idx_configurations_key ON public.configurations USING btree (key);


--
-- Name: idx_contact_evals_contact_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_contact_evals_contact_id ON public.contact_evals USING btree (contact_id);


--
-- Name: idx_contact_sources_contact_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_contact_sources_contact_id ON public.contact_sources USING btree (contact_id);


--
-- Name: idx_contact_sources_source_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_contact_sources_source_id ON public.contact_sources USING btree (source_id);


--
-- Name: idx_contact_tags_contact_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_contact_tags_contact_id ON public.contact_tags USING btree (contact_id);


--
-- Name: idx_contact_tags_tag_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_contact_tags_tag_id ON public.contact_tags USING btree (tag_id);


--
-- Name: idx_dispatch_analysis_transcription_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_dispatch_analysis_transcription_id ON public.dispatch_analysis USING btree (transcription_id);


--
-- Name: idx_dispatch_bulletpoint_analysis_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_dispatch_bulletpoint_analysis_id ON public.dispatch_bulletpoint USING btree (analysis_id);


--
-- Name: idx_dispatch_extracted_entity_bulletpoint_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_dispatch_extracted_entity_bulletpoint_id ON public.dispatch_extracted_entity USING btree (bulletpoint_id);


--
-- Name: idx_dispatch_transcription_audio_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_dispatch_transcription_audio_id ON public.dispatch_transcription USING btree (audio_id);


--
-- Name: idx_entities_type; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_entities_type ON public.entities USING btree (type);


--
-- Name: idx_entity_references_entity_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_entity_references_entity_id ON public.entity_references USING btree (entity_id);


--
-- Name: idx_entity_references_source; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_entity_references_source ON public.entity_references USING btree (source_type, source_id);


--
-- Name: idx_entity_relationships_entity_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_entity_relationships_entity_id ON public.entity_relationships USING btree (entity_id);


--
-- Name: idx_entity_relationships_related_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_entity_relationships_related_id ON public.entity_relationships USING btree (related_id);


--
-- Name: idx_entity_relationships_types; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_entity_relationships_types ON public.entity_relationships USING btree (related_type, relationship_type);


--
-- Name: idx_message_text_search; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_message_text_search ON public.message_text_representation USING gin (search_vector);


--
-- Name: idx_messages_body_gin; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_messages_body_gin ON public.messages USING gin (to_tsvector('english'::regconfig, body));


--
-- Name: idx_messages_created_at; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_messages_created_at ON public.messages_old USING btree (created_at);


--
-- Name: idx_messages_edit_history_message_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_messages_edit_history_message_id ON public.messages_edit_history USING btree (message_id);


--
-- Name: idx_messages_event_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_messages_event_id ON public.messages USING btree (event_id);


--
-- Name: idx_messages_external_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_messages_external_id ON public.messages_old USING btree (external_id);


--
-- Name: idx_messages_mentions_contact_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_messages_mentions_contact_id ON public.messages_mentions USING btree (contact_id);


--
-- Name: idx_messages_old_room_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_messages_old_room_id ON public.messages_old USING btree (room_id);


--
-- Name: idx_messages_relations_target; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_messages_relations_target ON public.messages_relations USING btree (target_event_id);


--
-- Name: idx_messages_room_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_messages_room_id ON public.messages USING btree (room_id);


--
-- Name: idx_messages_sender_contact_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_messages_sender_contact_id ON public.messages USING btree (sender_contact_id);


--
-- Name: idx_room_participants_room_contact; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_room_participants_room_contact ON public.room_participants USING btree (room_id, contact_id);


--
-- Name: idx_sessions_first_date_time; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_sessions_first_date_time ON public.sessions USING btree (first_date_time);


--
-- Name: idx_sessions_last_date_time; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_sessions_last_date_time ON public.sessions USING btree (last_date_time);


--
-- Name: idx_sessions_room_id; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_sessions_room_id ON public.sessions USING btree (room_id);


--
-- Name: idx_social_posts_created_at; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX idx_social_posts_created_at ON public.social_posts USING btree (created_at DESC NULLS LAST);


--
-- Name: room_state_active_idx; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX room_state_active_idx ON public.room_state USING btree (last_activity DESC) WHERE (is_hidden = false);


--
-- Name: room_state_last_activity_idx; Type: INDEX; Schema: public; Owner: gardener
--

CREATE INDEX room_state_last_activity_idx ON public.room_state USING btree (last_activity DESC);


--
-- Name: contacts after_contact_insert; Type: TRIGGER; Schema: public; Owner: gardener
--

CREATE TRIGGER after_contact_insert AFTER INSERT ON public.contacts FOR EACH ROW EXECUTE FUNCTION public.create_contact_entity();


--
-- Name: rooms after_room_insert; Type: TRIGGER; Schema: public; Owner: gardener
--

CREATE TRIGGER after_room_insert AFTER INSERT ON public.rooms FOR EACH ROW EXECUTE FUNCTION public.create_room_entity();


--
-- Name: contacts before_contact_delete; Type: TRIGGER; Schema: public; Owner: gardener
--

CREATE TRIGGER before_contact_delete BEFORE DELETE ON public.contacts FOR EACH ROW EXECUTE FUNCTION public.delete_contact_entity();


--
-- Name: message_text_representation message_text_search_update_trigger; Type: TRIGGER; Schema: public; Owner: gardener
--

CREATE TRIGGER message_text_search_update_trigger BEFORE INSERT OR UPDATE ON public.message_text_representation FOR EACH ROW EXECUTE FUNCTION public.message_text_search_update();


--
-- Name: bookmarks new_bookmark_trigger; Type: TRIGGER; Schema: public; Owner: gardener
--

CREATE TRIGGER new_bookmark_trigger AFTER INSERT ON public.bookmarks FOR EACH ROW EXECUTE FUNCTION public.notify_new_bookmark();


--
-- Name: raw_messages new_raw_message_to_message_trigger; Type: TRIGGER; Schema: public; Owner: gardener
--

CREATE TRIGGER new_raw_message_to_message_trigger AFTER INSERT ON public.raw_messages FOR EACH ROW EXECUTE FUNCTION public.process_raw_message_to_message();


--
-- Name: items notify_new_item_trigger; Type: TRIGGER; Schema: public; Owner: gardener
--

CREATE TRIGGER notify_new_item_trigger AFTER INSERT ON public.items FOR EACH ROW EXECUTE FUNCTION public.notify_new_item();


--
-- Name: messages process_new_message_into_session_trigger; Type: TRIGGER; Schema: public; Owner: gardener
--

CREATE TRIGGER process_new_message_into_session_trigger AFTER INSERT ON public.messages FOR EACH ROW EXECUTE FUNCTION public.process_new_message_into_session();


--
-- Name: messages process_new_message_type_trigger; Type: TRIGGER; Schema: public; Owner: gardener
--

CREATE TRIGGER process_new_message_type_trigger AFTER INSERT ON public.messages FOR EACH ROW EXECUTE FUNCTION public.process_message_type_eval();


--
-- Name: messages trigger_after_message_insert; Type: TRIGGER; Schema: public; Owner: gardener
--

CREATE TRIGGER trigger_after_message_insert AFTER INSERT ON public.messages FOR EACH ROW EXECUTE FUNCTION public.process_new_message_into_session();


--
-- Name: items trigger_item_name_to_slug; Type: TRIGGER; Schema: public; Owner: gardener
--

CREATE TRIGGER trigger_item_name_to_slug BEFORE INSERT ON public.items FOR EACH ROW WHEN (((new.title IS NOT NULL) AND (new.slug IS NULL))) EXECUTE FUNCTION public.set_slug_from_name();


--
-- Name: items update_modified_before_update; Type: TRIGGER; Schema: public; Owner: gardener
--

CREATE TRIGGER update_modified_before_update BEFORE UPDATE ON public.items FOR EACH ROW EXECUTE FUNCTION public.update_modified_column();


--
-- Name: item_tags update_tag_last_activity_after_insert_or_delete; Type: TRIGGER; Schema: public; Owner: gardener
--

CREATE TRIGGER update_tag_last_activity_after_insert_or_delete AFTER INSERT OR DELETE ON public.item_tags FOR EACH ROW EXECUTE FUNCTION public.update_tag_last_activity();


--
-- Name: tags update_tag_modified_before_update; Type: TRIGGER; Schema: public; Owner: gardener
--

CREATE TRIGGER update_tag_modified_before_update BEFORE UPDATE ON public.tags FOR EACH ROW EXECUTE FUNCTION public.update_tag_modified_column();


--
-- Name: alicia_message alicia_message_conversation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.alicia_message
    ADD CONSTRAINT alicia_message_conversation_id_fkey FOREIGN KEY (conversation_id) REFERENCES public.alicia_conversations(id);


--
-- Name: alicia_message alicia_message_previous_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.alicia_message
    ADD CONSTRAINT alicia_message_previous_id_fkey FOREIGN KEY (previous_id) REFERENCES public.alicia_message(id);


--
-- Name: alicia_meta alicia_meta_conversation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.alicia_meta
    ADD CONSTRAINT alicia_meta_conversation_id_fkey FOREIGN KEY (conversation_id) REFERENCES public.alicia_conversations(id);


--
-- Name: alicia_meta alicia_meta_ref_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.alicia_meta
    ADD CONSTRAINT alicia_meta_ref_fkey FOREIGN KEY (ref) REFERENCES public.alicia_message(id);


--
-- Name: bookmark_category bookmark_category_bookmark_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.bookmark_category
    ADD CONSTRAINT bookmark_category_bookmark_id_fkey FOREIGN KEY (bookmark_id) REFERENCES public.bookmarks(bookmark_id) ON DELETE CASCADE;


--
-- Name: bookmark_category bookmark_category_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.bookmark_category
    ADD CONSTRAINT bookmark_category_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.categories(category_id) ON DELETE CASCADE;


--
-- Name: bookmark_content_references bookmark_content_references_bookmark_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.bookmark_content_references
    ADD CONSTRAINT bookmark_content_references_bookmark_id_fkey FOREIGN KEY (bookmark_id) REFERENCES public.bookmarks(bookmark_id);


--
-- Name: bookmark_evaluations bookmark_evaluations_bookmark_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.bookmark_evaluations
    ADD CONSTRAINT bookmark_evaluations_bookmark_id_fkey FOREIGN KEY (bookmark_id) REFERENCES public.bookmarks(bookmark_id) ON DELETE CASCADE;


--
-- Name: bookmark_sources bookmark_sources_bookmark_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.bookmark_sources
    ADD CONSTRAINT bookmark_sources_bookmark_id_fkey FOREIGN KEY (bookmark_id) REFERENCES public.bookmarks(bookmark_id) ON DELETE CASCADE;


--
-- Name: bookmark_titles bookmark_titles_bookmark_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.bookmark_titles
    ADD CONSTRAINT bookmark_titles_bookmark_id_fkey FOREIGN KEY (bookmark_id) REFERENCES public.bookmarks(bookmark_id) ON DELETE CASCADE;


--
-- Name: category_sources category_sources_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.category_sources
    ADD CONSTRAINT category_sources_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.categories(category_id) ON DELETE CASCADE;


--
-- Name: contact_evals contact_evals_contact_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contact_evals
    ADD CONSTRAINT contact_evals_contact_id_fkey FOREIGN KEY (contact_id) REFERENCES public.contacts(contact_id) ON DELETE CASCADE;


--
-- Name: contact_known_avatars contact_known_avatars_contact_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contact_known_avatars
    ADD CONSTRAINT contact_known_avatars_contact_id_fkey FOREIGN KEY (contact_id) REFERENCES public.contacts(contact_id);


--
-- Name: contact_known_names contact_known_names_contact_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contact_known_names
    ADD CONSTRAINT contact_known_names_contact_id_fkey FOREIGN KEY (contact_id) REFERENCES public.contacts(contact_id);


--
-- Name: contact_sources contact_sources_contact_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contact_sources
    ADD CONSTRAINT contact_sources_contact_id_fkey FOREIGN KEY (contact_id) REFERENCES public.contacts(contact_id) ON DELETE CASCADE;


--
-- Name: contact_stats contact_stats_contact_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contact_stats
    ADD CONSTRAINT contact_stats_contact_id_fkey FOREIGN KEY (contact_id) REFERENCES public.contacts(contact_id) ON DELETE CASCADE;


--
-- Name: contact_tags contact_tags_contact_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contact_tags
    ADD CONSTRAINT contact_tags_contact_id_fkey FOREIGN KEY (contact_id) REFERENCES public.contacts(contact_id) ON DELETE CASCADE;


--
-- Name: contact_tags contact_tags_tag_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.contact_tags
    ADD CONSTRAINT contact_tags_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES public.contact_tagnames(tag_id) ON DELETE CASCADE;


--
-- Name: dispatch_analysis dispatch_analysis_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_analysis
    ADD CONSTRAINT dispatch_analysis_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.dispatch_analysis(id) ON DELETE SET NULL;


--
-- Name: dispatch_analysis dispatch_analysis_transcription_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_analysis
    ADD CONSTRAINT dispatch_analysis_transcription_id_fkey FOREIGN KEY (transcription_id) REFERENCES public.dispatch_transcription(id) ON DELETE CASCADE;


--
-- Name: dispatch_bulletpoint dispatch_bulletpoint_analysis_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_bulletpoint
    ADD CONSTRAINT dispatch_bulletpoint_analysis_id_fkey FOREIGN KEY (analysis_id) REFERENCES public.dispatch_analysis(id) ON DELETE CASCADE;


--
-- Name: dispatch_entity_merge dispatch_entity_merge_entity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_entity_merge
    ADD CONSTRAINT dispatch_entity_merge_entity_id_fkey FOREIGN KEY (entity_id) REFERENCES public.dispatch_extracted_entity(id) ON DELETE CASCADE;


--
-- Name: dispatch_entity_merge dispatch_entity_merge_merged_entity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_entity_merge
    ADD CONSTRAINT dispatch_entity_merge_merged_entity_id_fkey FOREIGN KEY (merged_entity_id) REFERENCES public.dispatch_extracted_entity(id) ON DELETE CASCADE;


--
-- Name: dispatch_extracted_entity dispatch_extracted_entity_bulletpoint_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_extracted_entity
    ADD CONSTRAINT dispatch_extracted_entity_bulletpoint_id_fkey FOREIGN KEY (bulletpoint_id) REFERENCES public.dispatch_bulletpoint(id) ON DELETE CASCADE;


--
-- Name: dispatch_transcription dispatch_transcription_audio_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_transcription
    ADD CONSTRAINT dispatch_transcription_audio_id_fkey FOREIGN KEY (audio_id) REFERENCES public.dispatch_audio_raw(id) ON DELETE CASCADE;


--
-- Name: dispatch_transcription dispatch_transcription_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_transcription
    ADD CONSTRAINT dispatch_transcription_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.dispatch_transcription(id) ON DELETE SET NULL;


--
-- Name: entity_references entity_references_entity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.entity_references
    ADD CONSTRAINT entity_references_entity_id_fkey FOREIGN KEY (entity_id) REFERENCES public.entities(entity_id);


--
-- Name: entity_relationships entity_relationships_entity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.entity_relationships
    ADD CONSTRAINT entity_relationships_entity_id_fkey FOREIGN KEY (entity_id) REFERENCES public.entities(entity_id);


--
-- Name: dispatch_job fk_project_analysis; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_job
    ADD CONSTRAINT fk_project_analysis FOREIGN KEY (analysis_id) REFERENCES public.dispatch_analysis(id) ON DELETE SET NULL;


--
-- Name: dispatch_job fk_project_audio; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_job
    ADD CONSTRAINT fk_project_audio FOREIGN KEY (audio_id) REFERENCES public.dispatch_audio_raw(id) ON DELETE SET NULL;


--
-- Name: dispatch_job fk_project_transcription; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.dispatch_job
    ADD CONSTRAINT fk_project_transcription FOREIGN KEY (transcription_id) REFERENCES public.dispatch_transcription(id) ON DELETE SET NULL;


--
-- Name: http_responses http_responses_bookmark_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.http_responses
    ADD CONSTRAINT http_responses_bookmark_id_fkey FOREIGN KEY (bookmark_id) REFERENCES public.bookmarks(bookmark_id) ON DELETE CASCADE;


--
-- Name: item_semantic_index item_semantic_index_item_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.item_semantic_index
    ADD CONSTRAINT item_semantic_index_item_id_fkey FOREIGN KEY (item_id) REFERENCES public.items(id);


--
-- Name: item_tags item_tags_item_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.item_tags
    ADD CONSTRAINT item_tags_item_id_fkey FOREIGN KEY (item_id) REFERENCES public.items(id);


--
-- Name: item_tags item_tags_tag_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.item_tags
    ADD CONSTRAINT item_tags_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES public.tags(id);


--
-- Name: message_text_representation message_text_representation_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.message_text_representation
    ADD CONSTRAINT message_text_representation_message_id_fkey FOREIGN KEY (message_id) REFERENCES public.messages(message_id) ON DELETE CASCADE;


--
-- Name: messages_edit_history messages_edit_history_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages_edit_history
    ADD CONSTRAINT messages_edit_history_message_id_fkey FOREIGN KEY (message_id) REFERENCES public.messages(message_id) ON DELETE CASCADE;


--
-- Name: messages_media messages_media_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages_media
    ADD CONSTRAINT messages_media_message_id_fkey FOREIGN KEY (message_id) REFERENCES public.messages(message_id) ON DELETE CASCADE;


--
-- Name: messages_mentions messages_mentions_contact_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages_mentions
    ADD CONSTRAINT messages_mentions_contact_id_fkey FOREIGN KEY (contact_id) REFERENCES public.contacts(contact_id);


--
-- Name: messages_mentions messages_mentions_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages_mentions
    ADD CONSTRAINT messages_mentions_message_id_fkey FOREIGN KEY (message_id) REFERENCES public.messages(message_id) ON DELETE CASCADE;


--
-- Name: messages_reactions messages_reactions_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages_reactions
    ADD CONSTRAINT messages_reactions_message_id_fkey FOREIGN KEY (message_id) REFERENCES public.messages(message_id) ON DELETE CASCADE;


--
-- Name: messages_reactions messages_reactions_sender_contact_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages_reactions
    ADD CONSTRAINT messages_reactions_sender_contact_id_fkey FOREIGN KEY (sender_contact_id) REFERENCES public.contacts(contact_id);


--
-- Name: messages_relations messages_relations_source_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages_relations
    ADD CONSTRAINT messages_relations_source_message_id_fkey FOREIGN KEY (source_message_id) REFERENCES public.messages(message_id) ON DELETE CASCADE;


--
-- Name: messages_old messages_room_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages_old
    ADD CONSTRAINT messages_room_id_fkey FOREIGN KEY (room_id) REFERENCES public.rooms(room_id);


--
-- Name: messages messages_room_id_fkey1; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_room_id_fkey1 FOREIGN KEY (room_id) REFERENCES public.rooms(room_id);


--
-- Name: messages messages_sender_contact_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_sender_contact_id_fkey FOREIGN KEY (sender_contact_id) REFERENCES public.contacts(contact_id);


--
-- Name: messages_old messages_sender_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.messages_old
    ADD CONSTRAINT messages_sender_id_fkey FOREIGN KEY (sender_id) REFERENCES public.contacts(contact_id);


--
-- Name: processed_contents processed_contents_bookmark_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.processed_contents
    ADD CONSTRAINT processed_contents_bookmark_id_fkey FOREIGN KEY (bookmark_id) REFERENCES public.bookmarks(bookmark_id) ON DELETE CASCADE;


--
-- Name: room_known_avatars room_known_avatars_room_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.room_known_avatars
    ADD CONSTRAINT room_known_avatars_room_id_fkey FOREIGN KEY (room_id) REFERENCES public.rooms(room_id);


--
-- Name: room_known_names room_known_names_room_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.room_known_names
    ADD CONSTRAINT room_known_names_room_id_fkey FOREIGN KEY (room_id) REFERENCES public.rooms(room_id);


--
-- Name: room_participants room_participants_contact_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.room_participants
    ADD CONSTRAINT room_participants_contact_id_fkey FOREIGN KEY (contact_id) REFERENCES public.contacts(contact_id);


--
-- Name: room_participants room_participants_room_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.room_participants
    ADD CONSTRAINT room_participants_room_id_fkey FOREIGN KEY (room_id) REFERENCES public.rooms(room_id);


--
-- Name: room_state room_state_last_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.room_state
    ADD CONSTRAINT room_state_last_message_id_fkey FOREIGN KEY (last_message_id) REFERENCES public.messages(message_id) ON DELETE SET NULL;


--
-- Name: room_state room_state_room_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.room_state
    ADD CONSTRAINT room_state_room_id_fkey FOREIGN KEY (room_id) REFERENCES public.rooms(room_id);


--
-- Name: session_message session_message_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.session_message
    ADD CONSTRAINT session_message_session_id_fkey FOREIGN KEY (session_id) REFERENCES public.sessions(session_id) ON DELETE CASCADE;


--
-- Name: session_summaries session_summaries_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.session_summaries
    ADD CONSTRAINT session_summaries_session_id_fkey FOREIGN KEY (session_id) REFERENCES public.sessions(session_id) ON DELETE CASCADE;


--
-- Name: sessions sessions_room_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: gardener
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_room_id_fkey FOREIGN KEY (room_id) REFERENCES public.rooms(room_id);


--
-- Name: electric_publication_default; Type: PUBLICATION; Schema: -; Owner: gardener
--

CREATE PUBLICATION electric_publication_default WITH (publish = 'insert, update, delete, truncate');


ALTER PUBLICATION electric_publication_default OWNER TO gardener;

--
-- Name: geek; Type: PUBLICATION; Schema: -; Owner: postgres
--

CREATE PUBLICATION geek FOR ALL TABLES WITH (publish = 'insert, update, delete, truncate');


ALTER PUBLICATION geek OWNER TO postgres;

--
-- Name: electric_publication_default contact_evals; Type: PUBLICATION TABLE; Schema: public; Owner: gardener
--

ALTER PUBLICATION electric_publication_default ADD TABLE ONLY public.contact_evals;


--
-- Name: electric_publication_default contact_known_avatars; Type: PUBLICATION TABLE; Schema: public; Owner: gardener
--

ALTER PUBLICATION electric_publication_default ADD TABLE ONLY public.contact_known_avatars;


--
-- Name: electric_publication_default contact_known_names; Type: PUBLICATION TABLE; Schema: public; Owner: gardener
--

ALTER PUBLICATION electric_publication_default ADD TABLE ONLY public.contact_known_names;


--
-- Name: electric_publication_default contact_stats; Type: PUBLICATION TABLE; Schema: public; Owner: gardener
--

ALTER PUBLICATION electric_publication_default ADD TABLE ONLY public.contact_stats;


--
-- Name: electric_publication_default contact_tagnames; Type: PUBLICATION TABLE; Schema: public; Owner: gardener
--

ALTER PUBLICATION electric_publication_default ADD TABLE ONLY public.contact_tagnames;


--
-- Name: electric_publication_default contact_tags; Type: PUBLICATION TABLE; Schema: public; Owner: gardener
--

ALTER PUBLICATION electric_publication_default ADD TABLE ONLY public.contact_tags;


--
-- Name: electric_publication_default contacts; Type: PUBLICATION TABLE; Schema: public; Owner: gardener
--

ALTER PUBLICATION electric_publication_default ADD TABLE ONLY public.contacts;


--
-- Name: electric_publication_default messages; Type: PUBLICATION TABLE; Schema: public; Owner: gardener
--

ALTER PUBLICATION electric_publication_default ADD TABLE ONLY public.messages;


--
-- Name: electric_publication_default messages_media; Type: PUBLICATION TABLE; Schema: public; Owner: gardener
--

ALTER PUBLICATION electric_publication_default ADD TABLE ONLY public.messages_media;


--
-- Name: electric_publication_default messages_mentions; Type: PUBLICATION TABLE; Schema: public; Owner: gardener
--

ALTER PUBLICATION electric_publication_default ADD TABLE ONLY public.messages_mentions;


--
-- Name: electric_publication_default messages_reactions; Type: PUBLICATION TABLE; Schema: public; Owner: gardener
--

ALTER PUBLICATION electric_publication_default ADD TABLE ONLY public.messages_reactions;


--
-- Name: electric_publication_default messages_relations; Type: PUBLICATION TABLE; Schema: public; Owner: gardener
--

ALTER PUBLICATION electric_publication_default ADD TABLE ONLY public.messages_relations;


--
-- Name: electric_publication_default room_participants; Type: PUBLICATION TABLE; Schema: public; Owner: gardener
--

ALTER PUBLICATION electric_publication_default ADD TABLE ONLY public.room_participants;


--
-- Name: electric_publication_default rooms; Type: PUBLICATION TABLE; Schema: public; Owner: gardener
--

ALTER PUBLICATION electric_publication_default ADD TABLE ONLY public.rooms;


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE USAGE ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO PUBLIC;
GRANT ALL ON SCHEMA public TO repl_garden;


--
-- Name: TABLE alicia_conversations; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.alicia_conversations TO repl_garden;


--
-- Name: TABLE alicia_message; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.alicia_message TO repl_garden;


--
-- Name: TABLE alicia_meta; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.alicia_meta TO repl_garden;


--
-- Name: TABLE bookmark_category; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.bookmark_category TO repl_garden;


--
-- Name: TABLE bookmark_content_references; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.bookmark_content_references TO repl_garden;


--
-- Name: TABLE bookmark_evaluations; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.bookmark_evaluations TO repl_garden;


--
-- Name: TABLE bookmark_sources; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.bookmark_sources TO repl_garden;


--
-- Name: TABLE bookmark_titles; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.bookmark_titles TO repl_garden;


--
-- Name: TABLE bookmarks; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.bookmarks TO repl_garden;


--
-- Name: TABLE browser_history; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.browser_history TO repl_garden;


--
-- Name: SEQUENCE browser_history_id_seq; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON SEQUENCE public.browser_history_id_seq TO repl_garden;


--
-- Name: TABLE categories; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.categories TO repl_garden;


--
-- Name: TABLE category_sources; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.category_sources TO repl_garden;


--
-- Name: TABLE configurations; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.configurations TO repl_garden;


--
-- Name: TABLE contact_evals; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.contact_evals TO repl_garden;


--
-- Name: TABLE contact_known_avatars; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.contact_known_avatars TO repl_garden;


--
-- Name: TABLE contact_known_names; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.contact_known_names TO repl_garden;


--
-- Name: TABLE contact_sources; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.contact_sources TO repl_garden;


--
-- Name: TABLE contact_stats; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.contact_stats TO repl_garden;


--
-- Name: TABLE contact_tagnames; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.contact_tagnames TO repl_garden;


--
-- Name: TABLE contact_tags; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.contact_tags TO repl_garden;


--
-- Name: TABLE contacts; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.contacts TO repl_garden;


--
-- Name: TABLE dispatch_analysis; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.dispatch_analysis TO repl_garden;


--
-- Name: SEQUENCE dispatch_analysis_id_seq; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON SEQUENCE public.dispatch_analysis_id_seq TO repl_garden;


--
-- Name: TABLE dispatch_audio_raw; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.dispatch_audio_raw TO repl_garden;


--
-- Name: SEQUENCE dispatch_audio_raw_id_seq; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON SEQUENCE public.dispatch_audio_raw_id_seq TO repl_garden;


--
-- Name: TABLE dispatch_bulletpoint; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.dispatch_bulletpoint TO repl_garden;


--
-- Name: SEQUENCE dispatch_bulletpoint_id_seq; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON SEQUENCE public.dispatch_bulletpoint_id_seq TO repl_garden;


--
-- Name: TABLE dispatch_entity_merge; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.dispatch_entity_merge TO repl_garden;


--
-- Name: SEQUENCE dispatch_entity_merge_id_seq; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON SEQUENCE public.dispatch_entity_merge_id_seq TO repl_garden;


--
-- Name: TABLE dispatch_extracted_entity; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.dispatch_extracted_entity TO repl_garden;


--
-- Name: SEQUENCE dispatch_extracted_entity_id_seq; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON SEQUENCE public.dispatch_extracted_entity_id_seq TO repl_garden;


--
-- Name: TABLE dispatch_job; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.dispatch_job TO repl_garden;


--
-- Name: SEQUENCE dispatch_job_id_seq; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON SEQUENCE public.dispatch_job_id_seq TO repl_garden;


--
-- Name: TABLE dispatch_transcription; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.dispatch_transcription TO repl_garden;


--
-- Name: SEQUENCE dispatch_transcription_id_seq; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON SEQUENCE public.dispatch_transcription_id_seq TO repl_garden;


--
-- Name: TABLE entities; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.entities TO repl_garden;


--
-- Name: TABLE entity_references; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.entity_references TO repl_garden;


--
-- Name: TABLE entity_relationships; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.entity_relationships TO repl_garden;


--
-- Name: TABLE item_tags; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.item_tags TO repl_garden;


--
-- Name: TABLE items; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.items TO repl_garden;


--
-- Name: TABLE tags; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.tags TO repl_garden;


--
-- Name: TABLE full_items; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.full_items TO repl_garden;


--
-- Name: TABLE http_responses; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.http_responses TO repl_garden;


--
-- Name: TABLE item_semantic_index; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.item_semantic_index TO repl_garden;


--
-- Name: TABLE message_text_representation; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.message_text_representation TO repl_garden;


--
-- Name: TABLE messages_old; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.messages_old TO repl_garden;


--
-- Name: TABLE rooms; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.rooms TO repl_garden;


--
-- Name: TABLE message_view; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.message_view TO repl_garden;


--
-- Name: TABLE messages; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.messages TO repl_garden;


--
-- Name: TABLE messages_edit_history; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.messages_edit_history TO repl_garden;


--
-- Name: TABLE messages_media; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.messages_media TO repl_garden;


--
-- Name: TABLE messages_mentions; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.messages_mentions TO repl_garden;


--
-- Name: TABLE messages_reactions; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.messages_reactions TO repl_garden;


--
-- Name: TABLE messages_relations; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.messages_relations TO repl_garden;


--
-- Name: TABLE observations; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.observations TO repl_garden;


--
-- Name: TABLE processed_contents; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.processed_contents TO repl_garden;


--
-- Name: TABLE raw_messages; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.raw_messages TO repl_garden;


--
-- Name: TABLE room_known_names; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.room_known_names TO repl_garden;


--
-- Name: TABLE room_participants; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.room_participants TO repl_garden;


--
-- Name: TABLE session_message; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.session_message TO repl_garden;


--
-- Name: TABLE session_summaries; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.session_summaries TO repl_garden;


--
-- Name: TABLE sessions; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.sessions TO repl_garden;


--
-- Name: TABLE social_posts; Type: ACL; Schema: public; Owner: gardener
--

GRANT ALL ON TABLE public.social_posts TO repl_garden;


--
-- PostgreSQL database dump complete
--

