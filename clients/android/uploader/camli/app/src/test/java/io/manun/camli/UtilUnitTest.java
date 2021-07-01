package io.manun.camli;

import android.net.Uri;

import org.junit.Assert;
import org.junit.Test;

import java.io.ByteArrayInputStream;
import java.io.File;
import java.io.FileDescriptor;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.net.URL;

public class UtilUnitTest {

    @Test
    public void slurp_CorrectInputStream_ReturnsString() throws IOException {
        InputStream in = new ByteArrayInputStream("test data".getBytes());
        String actual = Util.slurp(in);
        Assert.assertEquals("test data", actual);
    }

    @Test
    public void getSha1_CorrectFileDescriptor_ReturnsSha1() throws Exception {
        assert this.getClass().getClassLoader() != null;
        URL resource = this.getClass().getClassLoader().getResource("test.jpg");
        assert resource != null;
        File file = new File(resource.toURI());
        FileInputStream inputStream = new FileInputStream(file);
        FileDescriptor fd = inputStream.getFD();
        String actual = Util.getSha1(fd);
        Assert.assertEquals("c87f092bc124685bf754caf1f370c82c7d3fccc3", actual);
    }

}