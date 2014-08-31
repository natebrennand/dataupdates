--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET search_path = public, pg_catalog;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: courses_add_info; Type: TABLE; Schema: public; Owner: adicu; Tablespace: 
--

CREATE TABLE courses_add_info (
    coursefull character varying(32),
    globalcore boolean
);


ALTER TABLE public.courses_add_info OWNER TO adicu;

--
-- Name: courses_t; Type: TABLE; Schema: public; Owner: adicu; Tablespace: 
--

CREATE TABLE courses_t (
    term character varying(32),
    course character varying(32),
    prefixname character varying(32),
    divisioncode character varying(32),
    divisionname character varying(64),
    campuscode character varying(32),
    campusname character varying(32),
    schoolcode character varying(32),
    schoolname character varying(64),
    departmentcode character varying(32),
    departmentname character varying(64),
    subtermcode character varying(32),
    subtermname character varying(64),
    callnumber integer,
    numenrolled integer,
    maxsize integer,
    enrollmentstatus character varying(32),
    numfixedunits integer,
    minunits integer,
    maxunits integer,
    coursetitle character varying(64),
    coursesubtitle character varying(64),
    typecode character varying(32),
    typename character varying(32),
    approval character varying(32),
    bulletinflags character varying(32),
    classnotes character varying(64),
    meetson1 character varying(32),
    starttime1 time without time zone,
    endtime1 time without time zone,
    building1 character varying(32),
    room1 character varying(32),
    meetson2 character varying(32),
    starttime2 time without time zone,
    endtime2 time without time zone,
    building2 character varying(32),
    room2 character varying(32),
    meetson3 character varying(32),
    starttime3 time without time zone,
    endtime3 time without time zone,
    building3 character varying(32),
    room3 character varying(32),
    meetson4 character varying(32),
    starttime4 time without time zone,
    endtime4 time without time zone,
    building4 character varying(32),
    room4 character varying(32),
    meetson5 character varying(32),
    starttime5 time without time zone,
    endtime5 time without time zone,
    building5 character varying(32),
    room5 character varying(32),
    meetson6 character varying(32),
    starttime6 time without time zone,
    endtime6 time without time zone,
    building6 character varying(32),
    room6 character varying(32),
    meets1 character varying(64),
    meets2 character varying(64),
    meets3 character varying(64),
    meets4 character varying(64),
    meets5 character varying(64),
    meets6 character varying(64),
    instructor1name character varying(32),
    instructor2name character varying(32),
    instructor3name character varying(32),
    instructor4name character varying(32),
    prefixlongname character varying(32),
    exammeetson character varying(32),
    examstarttime time without time zone,
    examendtime time without time zone,
    exambuilding character varying(32),
    examroom character varying(32),
    exammeet character varying(64),
    examdate character varying(32),
    chargemsg1 character varying(32),
    chargeamt1 character varying(32),
    chargemsg2 character varying(32),
    chargeamt2 character varying(32),
    description text
);


ALTER TABLE public.courses_t OWNER TO adicu;

--
-- Name: courses_v2_t; Type: TABLE; Schema: public; Owner: adicu; Tablespace: 
--

CREATE TABLE courses_v2_t (
    course character varying(32) NOT NULL,
    coursefull character varying(32),
    prefixname character varying(32),
    divisioncode character varying(32),
    divisionname character varying(64),
    schoolcode character varying(32),
    schoolname character varying(64),
    departmentcode character varying(32),
    departmentname character varying(64),
    subtermcode character varying(32),
    subtermname character varying(64),
    enrollmentstatus character varying(32),
    numfixedunits integer,
    minunits integer,
    maxunits integer,
    coursetitle character varying(64),
    coursesubtitle character varying(64),
    approval character varying(32),
    bulletinflags character varying(32),
    classnotes character varying(64),
    prefixlongname character varying(32),
    description text
);


ALTER TABLE public.courses_v2_t OWNER TO adicu;

--
-- Name: housing_amenities_t; Type: TABLE; Schema: public; Owner: adicu; Tablespace: 
--

CREATE TABLE housing_amenities_t (
    building character varying(32),
    apartmentstyle boolean,
    suitestyle boolean,
    corridorstyle boolean,
    privatebathroom boolean,
    semiprivatebathroom boolean,
    sharedbathroom boolean,
    privatekitchen boolean,
    semiprivatekitchen boolean,
    sharedkitchen boolean,
    lounge character varying(32)
);


ALTER TABLE public.housing_amenities_t OWNER TO adicu;

--
-- Name: housing_t; Type: TABLE; Schema: public; Owner: adicu; Tablespace: 
--

CREATE TABLE housing_t (
    roomlocationarea character varying(32),
    residentialarea character varying(32),
    roomlocation character varying(32),
    roomlocationsection character varying(32),
    roomlocationfloorsuite character varying(32),
    issuite boolean,
    floorsuitewebdescription character varying(32),
    room character varying(32),
    roomarea integer,
    roomspace character varying(32),
    roomtype character varying(32),
    ay1213rsstatus character varying(32),
    pointvalue double precision,
    lotterynumber integer
);


ALTER TABLE public.housing_t OWNER TO adicu;

--
-- Name: sections_v2_t; Type: TABLE; Schema: public; Owner: adicu; Tablespace: 
--

CREATE TABLE sections_v2_t (
    callnumber integer,
    sectionfull character varying(32),
    bulletinurl character varying(32),
    course character varying(32),
    term character varying(32),
    numenrolled integer,
    maxsize integer,
    typecode character varying(32),
    typename character varying(32),
    meets1 character varying(64),
    meets2 character varying(64),
    meets3 character varying(64),
    meets4 character varying(64),
    meets5 character varying(64),
    meets6 character varying(64),
    meetson1 character varying(32),
    starttime1 time without time zone,
    endtime1 time without time zone,
    building1 character varying(32),
    room1 character varying(32),
    meetson2 character varying(32),
    starttime2 time without time zone,
    endtime2 time without time zone,
    building2 character varying(32),
    room2 character varying(32),
    meetson3 character varying(32),
    starttime3 time without time zone,
    endtime3 time without time zone,
    building3 character varying(32),
    room3 character varying(32),
    meetson4 character varying(32),
    starttime4 time without time zone,
    endtime4 time without time zone,
    building4 character varying(32),
    room4 character varying(32),
    meetson5 character varying(32),
    starttime5 time without time zone,
    endtime5 time without time zone,
    building5 character varying(32),
    room5 character varying(32),
    meetson6 character varying(32),
    starttime6 time without time zone,
    endtime6 time without time zone,
    building6 character varying(32),
    room6 character varying(32),
    exammeetson character varying(32),
    examstarttime time without time zone,
    examendtime time without time zone,
    exambuilding character varying(32),
    examroom character varying(32),
    exammeet character varying(64),
    examdate character varying(32),
    instructor1name character varying(32),
    instructor2name character varying(32),
    instructor3name character varying(32),
    instructor4name character varying(32),
    campuscode character varying(32),
    campusname character varying(32)
);


ALTER TABLE public.sections_v2_t OWNER TO adicu;

--
-- Name: users_t; Type: TABLE; Schema: public; Owner: adicu; Tablespace: 
--

CREATE TABLE users_t (
    email character varying(64) NOT NULL,
    token character varying(32) NOT NULL,
    name character varying(64) NOT NULL
);


ALTER TABLE public.users_t OWNER TO adicu;

--
-- Name: courses_v2_t_pkey; Type: CONSTRAINT; Schema: public; Owner: adicu; Tablespace: 
--

ALTER TABLE ONLY courses_v2_t
    ADD CONSTRAINT courses_v2_t_pkey PRIMARY KEY (course);


--
-- Name: users_t_email_key; Type: CONSTRAINT; Schema: public; Owner: adicu; Tablespace: 
--

ALTER TABLE ONLY users_t
    ADD CONSTRAINT users_t_email_key UNIQUE (email);


--
-- Name: sections_v2_t_course_fkey; Type: FK CONSTRAINT; Schema: public; Owner: adicu
--

ALTER TABLE ONLY sections_v2_t
    ADD CONSTRAINT sections_v2_t_course_fkey FOREIGN KEY (course) REFERENCES courses_v2_t(course);


--
-- Name: public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM postgres;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- Name: courses_add_info; Type: ACL; Schema: public; Owner: adicu
--

REVOKE ALL ON TABLE courses_add_info FROM PUBLIC;
REVOKE ALL ON TABLE courses_add_info FROM adicu;
GRANT ALL ON TABLE courses_add_info TO adicu;
GRANT SELECT ON TABLE courses_add_info TO adicu2;


--
-- Name: courses_t; Type: ACL; Schema: public; Owner: adicu
--

REVOKE ALL ON TABLE courses_t FROM PUBLIC;
REVOKE ALL ON TABLE courses_t FROM adicu;
GRANT ALL ON TABLE courses_t TO adicu;
GRANT SELECT ON TABLE courses_t TO adicu2;


--
-- Name: courses_v2_t; Type: ACL; Schema: public; Owner: adicu
--

REVOKE ALL ON TABLE courses_v2_t FROM PUBLIC;
REVOKE ALL ON TABLE courses_v2_t FROM adicu;
GRANT ALL ON TABLE courses_v2_t TO adicu;
GRANT SELECT ON TABLE courses_v2_t TO adicu2;


--
-- Name: housing_amenities_t; Type: ACL; Schema: public; Owner: adicu
--

REVOKE ALL ON TABLE housing_amenities_t FROM PUBLIC;
REVOKE ALL ON TABLE housing_amenities_t FROM adicu;
GRANT ALL ON TABLE housing_amenities_t TO adicu;
GRANT SELECT ON TABLE housing_amenities_t TO adicu2;


--
-- Name: housing_t; Type: ACL; Schema: public; Owner: adicu
--

REVOKE ALL ON TABLE housing_t FROM PUBLIC;
REVOKE ALL ON TABLE housing_t FROM adicu;
GRANT ALL ON TABLE housing_t TO adicu;
GRANT SELECT ON TABLE housing_t TO adicu2;


--
-- Name: sections_v2_t; Type: ACL; Schema: public; Owner: adicu
--

REVOKE ALL ON TABLE sections_v2_t FROM PUBLIC;
REVOKE ALL ON TABLE sections_v2_t FROM adicu;
GRANT ALL ON TABLE sections_v2_t TO adicu;
GRANT SELECT ON TABLE sections_v2_t TO adicu2;


--
-- Name: users_t; Type: ACL; Schema: public; Owner: adicu
--

REVOKE ALL ON TABLE users_t FROM PUBLIC;
REVOKE ALL ON TABLE users_t FROM adicu;
GRANT ALL ON TABLE users_t TO adicu;
GRANT SELECT ON TABLE users_t TO adicu2;


--
-- PostgreSQL database dump complete
--

