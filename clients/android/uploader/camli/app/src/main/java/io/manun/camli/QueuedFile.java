package io.manun.camli;

import android.net.Uri;

import java.util.Objects;

public class QueuedFile {
    private final String mContentName;
    private final Uri mUri;


    public QueuedFile(String sha1, Uri uri) {
        if (sha1 == null)
            throw new NullPointerException("sha1 == null");
        if (uri == null)
            throw new NullPointerException("uri == null");
        if (sha1.length() != 40)
            throw new IllegalArgumentException("unexpected sha1 length");
        mContentName = "sha1-" + sha1;
        mUri = uri;
    }

    public String getContentName() {
        return mContentName;
    }

    public Uri getUri() {
        return mUri;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        QueuedFile that = (QueuedFile) o;
        return Objects.equals(mContentName, that.mContentName) &&
                Objects.equals(mUri, that.mUri);
    }

    @Override
    public int hashCode() {
        return Objects.hash(mContentName, mUri);
    }
}
