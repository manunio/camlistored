package io.manun.camli;

import org.junit.Assert;
import org.junit.Test;

public class HostPortTest {

    @Test
    public void port_CorrectHostPort_ReturnsPort() {
        HostPort hp = new HostPort("localhost:3179");
        Assert.assertEquals(3179, hp.port());
    }

    @Test
    public void port_InCorrectHostPort_ReturnsZero() {
        HostPort hp = new HostPort("localhost::3179");
        Assert.assertEquals(0, hp.port());
    }

    @Test
    public void port_HostPortWithoutPort_ReturnsDefault() {
        HostPort hp = new HostPort("localhost");
        Assert.assertEquals(80, hp.port());
    }

    @Test
    public void host_CorrectHostPort_ReturnsHost() {
        HostPort hp = new HostPort("localhost:3179");
        Assert.assertEquals("localhost", hp.host());
    }

    @Test
    public void host_InCorrectHostPort_ReturnsZero() {
        HostPort hp = new HostPort("localhost::3179");
        Assert.assertNull(hp.host());
    }

    @Test
    public void host_HostPortWithoutPort_ReturnsDefault() {
        HostPort hp = new HostPort("localhost");
        Assert.assertEquals("localhost", hp.host());
    }

    @Test
    public void isValid_CorrectHostPort_ReturnsTrue() {
        HostPort hp = new HostPort("localhost");
        Assert.assertTrue(hp.isValid());
    }

    @Test
    public void isValid_InCorrectHostPort_ReturnsFalse() {
        HostPort hp = new HostPort("localhost::3179");
        Assert.assertFalse(hp.isValid());
    }

    @Test
    public void isValid_HostPortWithoutPort_ReturnsTrue() {
        HostPort hp = new HostPort("localhost");
        Assert.assertTrue(hp.isValid());
    }
}