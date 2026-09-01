"""Tests for MCP transport selection."""

import pytest

from mcp_config import (
    LOOPBACK_HOSTS,
    resolve_allowed_hosts,
    resolve_host,
    resolve_port,
    resolve_transport,
)


class TestResolveTransport:
    """Tests for resolve_transport()."""

    @pytest.mark.parametrize(
        ("value", "expected"),
        [
            (None, "stdio"),
            ("", "stdio"),
            ("   ", "stdio"),
            ("\t\n", "stdio"),
            ("  STDIO ", "stdio"),
            ("http", "streamable-http"),
            ("Http", "streamable-http"),
            ("streamable-http", "streamable-http"),
            ("streamable_http", "streamable-http"),
            ("sse", "sse"),
        ],
    )
    def test_valid_values(self, value, expected):
        assert resolve_transport(value) == expected

    def test_invalid_value_raises(self):
        with pytest.raises(ValueError, match="Invalid WHATSAPP_MCP_TRANSPORT"):
            resolve_transport("websocket")


class TestResolveHost:
    """Tests for resolve_host()."""

    @pytest.mark.parametrize(
        ("value", "expected"),
        [
            (None, "127.0.0.1"),
            ("", "127.0.0.1"),
            ("   ", "127.0.0.1"),
            ("\t\n", "127.0.0.1"),
            (" 127.0.0.1 ", "127.0.0.1"),
            ("0.0.0.0", "0.0.0.0"),
        ],
    )
    def test_values(self, value, expected):
        assert resolve_host(value) == expected


class TestResolvePort:
    """Tests for resolve_port()."""

    @pytest.mark.parametrize(
        ("value", "expected"),
        [
            (None, 8000),
            ("", 8000),
            ("   ", 8000),
            ("\t\n", 8000),
            ("9000", 9000),
            (" 9000 ", 9000),
            ("1", 1),
            ("65535", 65535),
        ],
    )
    def test_valid_values(self, value, expected):
        assert resolve_port(value) == expected

    def test_non_integer_raises(self):
        with pytest.raises(ValueError, match="Invalid WHATSAPP_MCP_PORT"):
            resolve_port("not-a-number")

    def test_out_of_range_raises(self):
        for value in ("0", "-1", "65536"):
            with pytest.raises(ValueError, match="Invalid WHATSAPP_MCP_PORT"):
                resolve_port(value)


class TestResolveAllowedHosts:
    """Tests for resolve_allowed_hosts()."""

    @pytest.mark.parametrize(
        ("value", "expected"),
        [
            (None, []),
            ("", []),
            ("   ", []),
            (",", []),
            (" , , ", []),
            ("example.internal:8000", ["example.internal:8000"]),
            ("example.internal:*", ["example.internal:*"]),
            (
                "a.internal:8000, b.internal:*",
                ["a.internal:8000", "b.internal:*"],
            ),
            (
                " a.internal:8000 ,, b.internal:8000 ",
                ["a.internal:8000", "b.internal:8000"],
            ),
        ],
    )
    def test_values(self, value, expected):
        assert resolve_allowed_hosts(value) == expected


class TestLoopbackHosts:
    """The set main.py checks before relaxing DNS-rebinding protection."""

    def test_covers_what_the_sdk_treats_as_loopback(self):
        assert LOOPBACK_HOSTS == {"127.0.0.1", "localhost", "::1"}

    @pytest.mark.parametrize("value", ["0.0.0.0", "example.internal", "::"])
    def test_excludes_non_loopback(self, value):
        assert resolve_host(value) not in LOOPBACK_HOSTS
